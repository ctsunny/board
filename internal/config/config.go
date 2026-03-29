package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

type GmailConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	FromEmail    string `json:"from_email"`
	AdminEmail   string `json:"admin_email"`
}

type Config struct {
	Port          int         `json:"port"`
	BasePath      string      `json:"base_path"`
	AdminUser     string      `json:"admin_user"`
	AdminPass     string      `json:"admin_pass"`
	AdminPassword string      `json:"admin_password"` // plain text, used only for displaying initial credentials; file is 0600 root-only
	JWTSecret     string      `json:"jwt_secret"`
	DBPath        string      `json:"db_path"`
	Domain        string      `json:"domain"`
	CertDir       string      `json:"cert_dir"`
	Gmail         GmailConfig `json:"gmail"`
	NotifyDays    []int       `json:"notify_days"`
	PingInterval  int         `json:"ping_interval"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func GenerateDefault() (*Config, error) {
	port, err := randomPort()
	if err != nil {
		return nil, err
	}
	suffix, err := randomHex(4)
	if err != nil {
		return nil, err
	}
	adminPass, err := randomHex(8)
	if err != nil {
		return nil, err
	}
	jwtSecret, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Config{
		Port:          port,
		BasePath:      "/mgmt-" + suffix,
		AdminUser:     "admin",
		AdminPass:     string(hashed),
		AdminPassword: adminPass,
		JWTSecret:     jwtSecret,
		DBPath:        "/var/lib/board/board.db",
		CertDir:       "/var/lib/board/certs",
		NotifyDays:    []int{7, 3, 1},
		PingInterval:  60,
	}, nil
}

func randomPort() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(55535))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 10000, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
