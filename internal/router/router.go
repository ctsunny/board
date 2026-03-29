package router

import (
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/ctsunny/board/internal/api"
	"github.com/ctsunny/board/internal/config"
	"github.com/ctsunny/board/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Setup registers all routes on the engine.
func Setup(r *gin.Engine, h *api.Handler, cfg *config.Config, database *gorm.DB, staticFiles fs.FS) {
	base := normalizeBasePath(cfg.BasePath)

	auth := middleware.CombinedAuth(cfg.JWTSecret, database)

	// API group
	v1 := r.Group(base + "/api/v1")
	{
		v1.POST("/auth/login", h.Login)

		protected := v1.Group("")
		protected.Use(auth)
		{
			protected.GET("/dashboard", h.Dashboard)

			// Customers
			protected.GET("/customers", h.ListCustomers)
			protected.POST("/customers", h.CreateCustomer)
			protected.GET("/customers/export", h.ExportCustomers)
			protected.POST("/customers/batch-delete", h.BatchDeleteCustomers)
			protected.POST("/customers/batch-renew", h.BatchRenewCustomers)
			protected.GET("/customers/:id", h.GetCustomer)
			protected.PUT("/customers/:id", h.UpdateCustomer)
			protected.DELETE("/customers/:id", h.DeleteCustomer)

			// Regions
			protected.GET("/regions", h.ListRegions)
			protected.POST("/regions", h.CreateRegion)
			protected.PUT("/regions/:id", h.UpdateRegion)
			protected.DELETE("/regions/:id", h.DeleteRegion)

			// Servers
			protected.GET("/servers", h.ListServers)
			protected.POST("/servers", h.CreateServer)
			protected.PUT("/servers/:id", h.UpdateServer)
			protected.DELETE("/servers/:id", h.DeleteServer)
			protected.POST("/servers/:id/ping", h.PingServerNow)

			// Routes
			protected.GET("/routes", h.ListRoutes)
			protected.POST("/routes", h.CreateRoute)
			protected.PUT("/routes/:id", h.UpdateRoute)
			protected.DELETE("/routes/:id", h.DeleteRoute)

			// Nodes
			protected.GET("/nodes", h.ListNodes)
			protected.POST("/nodes", h.CreateNode)
			protected.PUT("/nodes/:id", h.UpdateNode)
			protected.DELETE("/nodes/:id", h.DeleteNode)

			// Audit logs
			protected.GET("/audit-logs", h.ListAuditLogs)

			// API tokens
			protected.GET("/tokens", h.ListTokens)
			protected.POST("/tokens", h.CreateToken)
			protected.DELETE("/tokens/:id", h.DeleteToken)

			// Settings
			protected.GET("/settings", h.GetSettings)
			protected.PUT("/settings", h.UpdateSettings)
			protected.POST("/settings/gmail/test", h.TestGmailSend)
			protected.GET("/settings/gmail/auth-url", h.GmailAuthURL)
			protected.POST("/settings/gmail/callback", h.GmailCallback)
			protected.POST("/settings/telegram/test", h.TestTelegramSend)

			// System
			protected.GET("/system/version", h.SystemVersion)
			protected.POST("/system/update", h.SystemUpdate)
		}
	}

	// Serve Vue SPA
	if staticFiles != nil {
		// Pre-load and patch index.html with runtime config injection
		indexData, _ := fs.ReadFile(staticFiles, "index.html")
		if indexData != nil {
			inject := `<script>window.__BOARD_BASE__=` + strconv.Quote(base) + `;</script>`
			indexData = []byte(strings.Replace(string(indexData), "<head>", "<head>"+inject, 1))
		}

		fileServer := http.FileServer(http.FS(staticFiles))
		// Wrap with StripPrefix so the file server can resolve paths
		// when a base path (e.g. /mgmt-xxxx) is configured.
		fileHandler := http.StripPrefix(base, fileServer)

		r.NoRoute(func(c *gin.Context) {
			urlPath := c.Request.URL.Path
			if base != "" && urlPath == base {
				c.Redirect(http.StatusPermanentRedirect, base+"/")
				return
			}
			// API routes that are not found should return 404 JSON
			if strings.HasPrefix(urlPath, base+"/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			// Strip base path prefix, then leading slash to get the relative
			// path within the embedded static FS.
			trimmed := strings.TrimPrefix(urlPath, base)
			trimmed = strings.TrimPrefix(trimmed, "/")
			// Try to serve static asset file; on failure fall back to index.html (SPA)
			if trimmed != "" && trimmed != "index.html" {
				if _, err := staticFiles.Open(trimmed); err == nil {
					fileHandler.ServeHTTP(c.Writer, c.Request)
					return
				}
			}
			// SPA fallback: serve patched index.html
			if indexData == nil {
				c.Status(http.StatusNotFound)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
		})
	}
}

func normalizeBasePath(base string) string {
	base = strings.TrimSpace(base)
	if base == "" || base == "/" {
		return ""
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	return strings.TrimRight(base, "/")
}
