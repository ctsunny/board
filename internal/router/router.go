package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ctsunny/board/internal/api"
	"github.com/ctsunny/board/internal/config"
	"github.com/ctsunny/board/internal/middleware"
	"gorm.io/gorm"
)

// Setup registers all routes on the engine.
func Setup(r *gin.Engine, h *api.Handler, cfg *config.Config, database *gorm.DB, staticFiles fs.FS) {
	base := cfg.BasePath
	if base == "" {
		base = "/"
	}

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
			protected.GET("/settings/gmail/auth-url", h.GmailAuthURL)
			protected.POST("/settings/gmail/callback", h.GmailCallback)

			// System
			protected.GET("/system/version", h.SystemVersion)
			protected.POST("/system/update", h.SystemUpdate)
		}
	}

	// Serve Vue SPA
	if staticFiles != nil {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// API routes that are not found should return 404 JSON
			if strings.HasPrefix(path, base+"/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			// Try to serve the file directly
			f, err := staticFiles.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				f.Close()
				http.FileServer(http.FS(staticFiles)).ServeHTTP(c.Writer, c.Request)
				return
			}
			// SPA fallback: serve index.html
			indexData, err := fs.ReadFile(staticFiles, "index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
		})
	}
}
