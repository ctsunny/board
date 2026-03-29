package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ctsunny/board/internal/api"
	"github.com/ctsunny/board/internal/config"
	dbpkg "github.com/ctsunny/board/internal/db"
	"github.com/ctsunny/board/internal/router"
	"github.com/ctsunny/board/internal/services"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var webDist embed.FS

func main() {
	cfgPath := flag.String("config", "/etc/board/config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Config not found, generating defaults...")
			cfg, err = config.GenerateDefault()
			if err != nil {
				log.Fatalf("generate config: %v", err)
			}
			if err := config.Save(*cfgPath, cfg); err != nil {
				log.Printf("Warning: could not save config to %s: %v", *cfgPath, err)
			} else {
				log.Printf("Config saved to %s", *cfgPath)
			}
		} else {
			log.Fatalf("load config: %v", err)
		}
	}
	if reason, err := config.EnsureAdminCredentials(cfg); err != nil {
		log.Fatalf("ensure admin credentials: %v", err)
	} else if reason != "" {
		if err := config.Save(*cfgPath, cfg); err != nil {
			log.Printf("Warning: could not persist updated admin credentials to %s: %v", *cfgPath, err)
		} else {
			log.Printf("Config updated at %s: %s", *cfgPath, reason)
		}
	}

	// Init DB
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = "board.db"
	}
	if err := os.MkdirAll(dirOf(dbPath), 0o750); err != nil && !os.IsExist(err) {
		log.Printf("Warning: could not create db dir: %v", err)
	}
	database, err := dbpkg.Init(dbPath)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	// Services
	notifier := services.NewGmailNotifier(cfg.Gmail)
	services.StartExpiryScheduler(database, cfg, notifier)
	services.StartPingScheduler(database, cfg, notifier)

	// Static files sub-FS
	staticFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}

	// Gin engine
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	h := &api.Handler{
		DB:       database,
		Cfg:      cfg,
		Notifier: notifier,
		CfgPath:  *cfgPath,
	}
	router.Setup(engine, h, cfg, database, staticFS)

	// Start server
	if cfg.Domain != "" {
		startHTTPS(engine, cfg)
	} else {
		startHTTP(engine, cfg)
	}
}

func startHTTP(engine *gin.Engine, cfg *config.Config) {
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}
	log.Printf("Starting HTTP server on %s (base path: %s)", addr, cfg.BasePath)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()
	gracefulShutdown(srv)
}

func startHTTPS(engine *gin.Engine, cfg *config.Config) {
	tlsCfg, err := services.SetupTLS(cfg.Domain, cfg.CertDir)
	if err != nil {
		log.Fatalf("setup tls: %v", err)
	}
	srv := &http.Server{
		Addr:      ":443",
		Handler:   engine,
		TLSConfig: tlsCfg,
	}
	// Redirect HTTP → HTTPS
	go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := "https://" + cfg.Domain + r.RequestURI
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}))
	log.Printf("Starting HTTPS server for domain %s (base path: %s)", cfg.Domain, cfg.BasePath)
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen tls: %v", err)
		}
	}()
	gracefulShutdown(srv)
}

func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
