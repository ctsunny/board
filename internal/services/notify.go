package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ctsunny/board/internal/config"
	"github.com/ctsunny/board/internal/models"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const gmailScope = "https://www.googleapis.com/auth/gmail.send"

type GmailNotifier struct {
	cfg config.GmailConfig
}

func NewGmailNotifier(cfg config.GmailConfig) *GmailNotifier {
	return &GmailNotifier{cfg: cfg}
}

func (n *GmailNotifier) IsConfigured() bool {
	return n.cfg.RefreshToken != "" && n.cfg.ClientID != "" && n.cfg.ClientSecret != ""
}

func (n *GmailNotifier) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     n.cfg.ClientID,
		ClientSecret: n.cfg.ClientSecret,
		Scopes:       []string{gmailScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
}

func (n *GmailNotifier) httpClient(ctx context.Context) *http.Client {
	oauthCfg := n.oauthConfig()
	token := &oauth2.Token{
		RefreshToken: n.cfg.RefreshToken,
		Expiry:       time.Now().Add(-time.Second),
	}
	tokenSource := oauthCfg.TokenSource(ctx, token)
	return oauth2.NewClient(ctx, tokenSource)
}

func (n *GmailNotifier) SendEmail(to, subject, body string) error {
	if !n.IsConfigured() {
		return fmt.Errorf("gmail not configured")
	}
	ctx := context.Background()
	client := n.httpClient(ctx)

	rawMsg := buildRawEmail(n.cfg.FromEmail, to, subject, body)
	encoded := base64.URLEncoding.EncodeToString([]byte(rawMsg))

	payload := map[string]string{"raw": encoded}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gmail API error %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func buildRawEmail(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}

func (n *GmailNotifier) SendExpiryReminder(customer models.Customer, daysLeft int) error {
	subject := fmt.Sprintf("[Board] Customer %s expires in %d day(s)", customer.Name, daysLeft)
	body := fmt.Sprintf(
		"Customer expiry reminder\n\nName: %s\nContact: %s\nExpires: %s\nDays left: %d\n",
		customer.Name, customer.Contact, customer.ExpiresAt.Format("2006-01-02"), daysLeft,
	)
	return n.SendEmail(n.cfg.AdminEmail, subject, body)
}

func (n *GmailNotifier) SendServerDownAlert(server models.Server) error {
	subject := fmt.Sprintf("[Board] Server %s is OFFLINE", server.Name)
	body := fmt.Sprintf(
		"Server down alert\n\nName: %s\nIP: %s\nLocation: %s\nTime: %s\n",
		server.Name, server.IP, server.Location, time.Now().Format(time.RFC3339),
	)
	return n.SendEmail(n.cfg.AdminEmail, subject, body)
}

func StartExpiryScheduler(database *gorm.DB, cfg *config.Config, notifier *GmailNotifier) {
	c := cron.New()
	c.AddFunc("0 9 * * *", func() {
		checkAndNotifyExpiry(database, cfg, notifier)
	})
	c.Start()
}

func checkAndNotifyExpiry(database *gorm.DB, cfg *config.Config, notifier *GmailNotifier) {
	if !notifier.IsConfigured() {
		return
	}
	now := time.Now()
	for _, days := range cfg.NotifyDays {
		start := now.Add(time.Duration(days)*24*time.Hour - time.Minute)
		end := now.Add(time.Duration(days)*24*time.Hour + 23*time.Hour + 59*time.Minute)

		var customers []models.Customer
		database.Where("status = 'active' AND expires_at BETWEEN ? AND ?", start, end).Find(&customers)
		for _, c := range customers {
			err := notifier.SendExpiryReminder(c, days)
			status := "sent"
			errMsg := ""
			if err != nil {
				status = "failed"
				errMsg = err.Error()
			}
			database.Create(&models.NotificationLog{
				Type:           "expiry",
				RecipientEmail: cfg.Gmail.AdminEmail,
				Subject:        fmt.Sprintf("Customer %s expires in %d day(s)", c.Name, days),
				Status:         status,
				Error:          errMsg,
			})
		}
	}
}

func GetOAuthURL(clientID, clientSecret string) string {
	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{gmailScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	return oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func ExchangeCode(clientID, clientSecret, code string) (refreshToken string, email string, err error) {
	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{gmailScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}

	ctx := context.Background()
	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return "", "", err
	}

	// Fetch user email
	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?fields=email")
	if err == nil {
		defer resp.Body.Close()
		var info struct {
			Email string `json:"email"`
		}
		if err2 := json.NewDecoder(resp.Body).Decode(&info); err2 == nil {
			email = info.Email
		}
	}

	// Fallback: parse id_token if present
	if email == "" {
		if idToken, ok := token.Extra("id_token").(string); ok {
			parts := strings.Split(idToken, ".")
			if len(parts) == 3 {
				if decoded, err2 := base64.RawURLEncoding.DecodeString(parts[1]); err2 == nil {
					var payload struct {
						Email string `json:"email"`
					}
					if json.Unmarshal(decoded, &payload) == nil {
						email = payload.Email
					}
				}
			}
		}
	}

	return token.RefreshToken, email, nil
}

// fetchUserEmail via tokeninfo endpoint.
func fetchUserEmail(accessToken string) string {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(accessToken))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var info struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return ""
	}
	return info.Email
}
