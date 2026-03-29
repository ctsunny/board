package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctsunny/board/internal/config"
	dbpkg "github.com/ctsunny/board/internal/db"
	"github.com/ctsunny/board/internal/services"
	"github.com/gin-gonic/gin"
)

func TestSettingsSupportFrontendGmailFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	database, err := dbpkg.Init(filepath.Join(tmpDir, "board.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	cfg := &config.Config{
		NotifyDays:   []int{7, 3, 1},
		PingInterval: 60,
	}
	h := &Handler{
		DB:       database,
		Cfg:      cfg,
		Telegram: services.NewTelegramNotifierRef(&cfg.Telegram),
		CfgPath:  cfgPath,
	}

	updateRec := httptest.NewRecorder()
	updateCtx, _ := gin.CreateTestContext(updateRec)
	updateCtx.Request = httptest.NewRequest(
		http.MethodPut,
		"/settings",
		strings.NewReader(`{"gmail_client_id":"client-id","gmail_client_secret":"client-secret","gmail_admin_email":"admin@example.com","telegram_bot_token":"tg-token","telegram_chat_id":"123456"}`),
	)
	updateCtx.Request.Header.Set("Content-Type", "application/json")

	h.UpdateSettings(updateCtx)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d with body %s", http.StatusOK, updateRec.Code, updateRec.Body.String())
	}
	if got := h.Cfg.Gmail.ClientID; got != "client-id" {
		t.Fatalf("expected client id to be saved, got %q", got)
	}
	if got := h.Cfg.Gmail.ClientSecret; got != "client-secret" {
		t.Fatalf("expected client secret to be saved, got %q", got)
	}
	if got := h.Cfg.Gmail.AdminEmail; got != "admin@example.com" {
		t.Fatalf("expected admin email to be saved, got %q", got)
	}
	if got := h.Cfg.Telegram.BotToken; got != "tg-token" {
		t.Fatalf("expected telegram bot token to be saved, got %q", got)
	}
	if got := h.Cfg.Telegram.ChatID; got != "123456" {
		t.Fatalf("expected telegram chat id to be saved, got %q", got)
	}

	saved, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	if !strings.Contains(string(saved), `"client_id": "client-id"`) {
		t.Fatalf("expected saved config to contain client id, got %s", string(saved))
	}

	getRec := httptest.NewRecorder()
	getCtx, _ := gin.CreateTestContext(getRec)
	getCtx.Request = httptest.NewRequest(http.MethodGet, "/settings", nil)

	h.GetSettings(getCtx)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get status %d, got %d", http.StatusOK, getRec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(getRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := body["gmail_client_id"]; got != "client-id" {
		t.Fatalf("expected gmail_client_id to round-trip, got %#v", got)
	}
	if got := body["gmail_admin_email"]; got != "admin@example.com" {
		t.Fatalf("expected gmail_admin_email to round-trip, got %#v", got)
	}
	if got := body["gmail_configured"]; got != false {
		t.Fatalf("expected gmail_configured false before oauth callback, got %#v", got)
	}
	if got := body["telegram_bot_token"]; got != "tg-token" {
		t.Fatalf("expected telegram_bot_token to round-trip, got %#v", got)
	}
	if got := body["telegram_chat_id"]; got != "123456" {
		t.Fatalf("expected telegram_chat_id to round-trip, got %#v", got)
	}
	if got := body["telegram_configured"]; got != true {
		t.Fatalf("expected telegram_configured true after saving token/chat id, got %#v", got)
	}
}
