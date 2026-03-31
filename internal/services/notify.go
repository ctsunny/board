package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ctsunny/board/internal/config"
	"github.com/ctsunny/board/internal/models"
	"github.com/robfig/cron/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

const gmailScope = "https://www.googleapis.com/auth/gmail.send"

type GmailNotifier struct {
	cfg *config.GmailConfig
}

func NewGmailNotifier(cfg config.GmailConfig) *GmailNotifier {
	cfgCopy := cfg
	return &GmailNotifier{cfg: &cfgCopy}
}

func NewGmailNotifierRef(cfg *config.GmailConfig) *GmailNotifier {
	return &GmailNotifier{cfg: cfg}
}

func (n *GmailNotifier) IsConfigured() bool {
	return n != nil && n.cfg != nil && n.cfg.RefreshToken != "" && n.cfg.ClientID != "" && n.cfg.ClientSecret != ""
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

type TelegramNotifier struct {
	cfg *config.TelegramConfig
}

func NewTelegramNotifier(cfg config.TelegramConfig) *TelegramNotifier {
	cfgCopy := cfg
	return &TelegramNotifier{cfg: &cfgCopy}
}

func NewTelegramNotifierRef(cfg *config.TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{cfg: cfg}
}

func (n *TelegramNotifier) IsConfigured() bool {
	return n != nil && n.cfg != nil && strings.TrimSpace(n.cfg.BotToken) != "" && strings.TrimSpace(n.cfg.ChatID) != ""
}

func (n *TelegramNotifier) SendMessage(text string) error {
	if !n.IsConfigured() {
		return fmt.Errorf("telegram not configured")
	}
	payload := map[string]string{
		"chat_id": n.cfg.ChatID,
		"text":    text,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost,
		"https://api.telegram.org/bot"+n.cfg.BotToken+"/sendMessage",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func buildRawEmail(from, to, subject, body string) string {
	from = sanitizeHeaderValue(from)
	to = sanitizeHeaderValue(to)
	subject = mime.QEncoding.Encode("UTF-8", sanitizeHeaderValue(subject))

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

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

func formatExpiryReminderSubject(customer models.Customer) string {
	return fmt.Sprintf("到期提醒 %s %s", customer.Name, customer.ExpiresAt.Format("2006-01-02"))
}

func formatServerDownAlertSubject(server models.Server, alertAt time.Time) string {
	return fmt.Sprintf("离线提醒 %s %s", server.Name, alertAt.Format("2006-01-02"))
}

func (n *GmailNotifier) SendExpiryReminder(customer models.Customer, daysLeft int) error {
	subject := formatExpiryReminderSubject(customer)
	body := fmt.Sprintf(
		"到期提醒\n\n客户：%s\n联系方式：%s\n到期日期：%s\n剩余天数：%d\n请及时跟进续费。\n",
		customer.Name, customer.Contact, customer.ExpiresAt.Format("2006-01-02"), daysLeft,
	)
	return n.SendEmail(n.cfg.AdminEmail, subject, body)
}

func (n *GmailNotifier) SendServerDownAlert(server models.Server) error {
	alertAt := time.Now()
	subject := formatServerDownAlertSubject(server, alertAt)
	body := fmt.Sprintf(
		"离线提醒\n\n服务器：%s\nIP：%s\n位置：%s\n告警时间：%s\n请尽快检查服务器状态。\n",
		server.Name, server.IP, server.Location, alertAt.Format("2006-01-02 15:04:05"),
	)
	return n.SendEmail(n.cfg.AdminEmail, subject, body)
}

func (n *TelegramNotifier) SendExpiryReminder(customer models.Customer, daysLeft int) error {
	return n.SendMessage(fmt.Sprintf(
		"📅 客户到期提醒\n客户：%s\n联系方式：%s\n到期日：%s\n剩余天数：%d",
		customer.Name,
		customer.Contact,
		customer.ExpiresAt.Format("2006-01-02"),
		daysLeft,
	))
}

func (n *TelegramNotifier) SendNewCustomerAlert(customer models.Customer) error {
	return n.SendMessage(fmt.Sprintf(
		"🆕 新增客户提醒\n客户：%s\n联系方式：%s\n地区：%s\n线路：%s\n服务器：%s\n节点：%s",
		customer.Name,
		customer.Contact,
		customer.RegionName,
		customer.RouteName,
		customer.ServerName,
		customer.NodeName,
	))
}

func (n *TelegramNotifier) SendLoginAlert(username, ip string) error {
	return n.SendMessage(fmt.Sprintf(
		"🔐 登录提醒\n账号：%s\nIP：%s\n时间：%s",
		username,
		ip,
		time.Now().Format(time.RFC3339),
	))
}

func (n *TelegramNotifier) SendFailedLoginAlert(username, ip string, attempts int64) error {
	return n.SendMessage(fmt.Sprintf(
		"⚠️ 密码错误提醒\n账号：%s\nIP：%s\n失败次数：%d\n时间：%s",
		username,
		ip,
		attempts,
		time.Now().Format(time.RFC3339),
	))
}

func StartExpiryScheduler(database *gorm.DB, cfg *config.Config, notifier *GmailNotifier, telegram *TelegramNotifier) {
	c := cron.New()
	c.AddFunc("0 9 * * *", func() {
		checkAndNotifyExpiry(database, cfg, notifier, telegram)
	})
	c.Start()
}

func checkAndNotifyExpiry(database *gorm.DB, cfg *config.Config, notifier *GmailNotifier, telegram *TelegramNotifier) {
	if (notifier == nil || !notifier.IsConfigured()) && (telegram == nil || !telegram.IsConfigured()) {
		return
	}
	now := time.Now()
	for _, days := range cfg.NotifyDays {
		start := now.Add(time.Duration(days)*24*time.Hour - time.Minute)
		end := now.Add(time.Duration(days)*24*time.Hour + 23*time.Hour + 59*time.Minute)

		var customers []models.Customer
		database.Where("status = 'active' AND expires_at BETWEEN ? AND ?", start, end).Find(&customers)
		for _, customer := range customers {
			logNotification(database, models.NotificationLog{
				Type:           "expiry_email",
				RecipientEmail: cfg.Gmail.AdminEmail,
				Subject:        formatExpiryReminderSubject(customer),
			}, notifier != nil && notifier.IsConfigured(), func() error {
				return notifier.SendExpiryReminder(customer, days)
			})
			logNotification(database, models.NotificationLog{
				Type:    "expiry_tg",
				Subject: formatExpiryReminderSubject(customer),
			}, telegram != nil && telegram.IsConfigured(), func() error {
				return telegram.SendExpiryReminder(customer, days)
			})
		}
	}
}

func LogTelegramNotification(database *gorm.DB, telegram *TelegramNotifier, typ, subject string, send func() error) {
	logNotification(database, models.NotificationLog{
		Type:    typ,
		Subject: subject,
	}, telegram != nil && telegram.IsConfigured(), send)
}

func CountRecentFailedLogins(database *gorm.DB, username, ip string, since time.Time) int64 {
	var count int64
	database.Model(&models.AuditLog{}).
		Where("action = ? AND resource = ? AND operator = ? AND ip = ? AND created_at >= ?", "login_failed", "auth", username, ip, since).
		Count(&count)
	return count
}

func logNotification(database *gorm.DB, record models.NotificationLog, enabled bool, send func() error) {
	if !enabled {
		return
	}
	err := send()
	record.Status = "sent"
	if err != nil {
		record.Status = "failed"
		record.Error = err.Error()
	}
	database.Create(&record)
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
