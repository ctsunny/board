package services

import (
	"mime"
	"strings"
	"testing"
	"time"

	"github.com/ctsunny/board/internal/models"
)

func TestBuildRawEmailEncodesUTF8SubjectAndSanitizesHeaders(t *testing.T) {
	raw := buildRawEmail("from@example.com", "to@example.com", "到期提醒\r\n张三 2026-04-01", "正文")
	expectedSubject := mime.QEncoding.Encode("UTF-8", "到期提醒 张三 2026-04-01")

	if !strings.Contains(raw, "Subject: "+expectedSubject+"\r\n") {
		t.Fatalf("expected encoded UTF-8 subject, got %q", raw)
	}
	if strings.Contains(raw, "\r\n张三 2026-04-01") {
		t.Fatalf("expected subject header injection to be sanitized, got %q", raw)
	}
	if !strings.Contains(raw, expectedSubject) {
		t.Fatalf("expected sanitized subject value %q in raw email, got %q", expectedSubject, raw)
	}
}

func TestNotificationSubjectsUseActionNameDateFormat(t *testing.T) {
	customer := models.Customer{
		Name:      "某某客户",
		ExpiresAt: time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC),
	}
	server := models.Server{Name: "香港 50 111 -2"}
	alertAt := time.Date(2026, 3, 31, 6, 10, 0, 0, time.UTC)

	if got := formatExpiryReminderSubject(customer); got != "到期提醒 某某客户 2026-04-01" {
		t.Fatalf("unexpected expiry reminder subject: %q", got)
	}
	if got := formatServerDownAlertSubject(server, alertAt); got != "离线提醒 香港 50 111 -2 2026-03-31" {
		t.Fatalf("unexpected server down subject: %q", got)
	}
}
