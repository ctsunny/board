package services

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ctsunny/board/internal/config"
	"github.com/ctsunny/board/internal/models"
	"gorm.io/gorm"
)

var latencyRe = regexp.MustCompile(`time[=<](\d+(?:\.\d+)?)`)

// PingServer pings ip once and returns latency in ms, or -1 if unreachable.
func PingServer(ip string, timeout time.Duration) (int, error) {
	secs := int(timeout.Seconds())
	if secs < 1 {
		secs = 1
	}
	cmd := exec.Command("ping", "-c", "1", "-W", fmt.Sprintf("%d", secs), ip)
	out, err := cmd.Output()
	if err != nil {
		return -1, nil
	}
	output := string(out)
	m := latencyRe.FindStringSubmatch(output)
	if m == nil {
		return -1, nil
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(m[1]), 64)
	if err != nil {
		return -1, nil
	}
	return int(f), nil
}

// StartPingScheduler runs periodic pings against all servers and updates DB.
func StartPingScheduler(database *gorm.DB, cfg *config.Config, notifier *GmailNotifier) {
	interval := time.Duration(cfg.PingInterval) * time.Second
	if interval < time.Second {
		interval = 60 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			runPingCycle(database, notifier)
		}
	}()
}

func runPingCycle(database *gorm.DB, notifier *GmailNotifier) {
	var servers []models.Server
	database.Find(&servers)

	for _, srv := range servers {
		latency, _ := PingServer(srv.IP, 2*time.Second)
		prevStatus := srv.Status

		newStatus := "online"
		if latency < 0 {
			newStatus = "offline"
		}

		now := time.Now()
		database.Model(&srv).Updates(map[string]interface{}{
			"latency":      latency,
			"status":       newStatus,
			"last_ping_at": now,
		})

		// Alert if newly offline
		if prevStatus == "online" && newStatus == "offline" {
			if notifier != nil && notifier.IsConfigured() {
				srv.Status = newStatus
				err := notifier.SendServerDownAlert(srv)
				status := "sent"
				errMsg := ""
				if err != nil {
					status = "failed"
					errMsg = err.Error()
				}
				database.Create(&models.NotificationLog{
					Type:    "server_down",
					Subject: formatServerDownAlertSubject(srv, now),
					Status:  status,
					Error:   errMsg,
				})
			}
		}
	}
}
