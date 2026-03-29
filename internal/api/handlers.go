package api

import (
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ctsunny/board/internal/config"
	"github.com/ctsunny/board/internal/middleware"
	"github.com/ctsunny/board/internal/models"
	"github.com/ctsunny/board/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB       *gorm.DB
	Cfg      *config.Config
	Notifier *services.GmailNotifier
	CfgPath  string
}

// ---------- helpers ----------

type PageResult struct {
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Data    interface{} `json:"data"`
}

func pageParams(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 20
	}
	return
}

func offset(page, perPage int) int {
	return (page - 1) * perPage
}

func clientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return c.ClientIP()
}

func (h *Handler) audit(action, resource string, resourceID uint, detail, ip string) {
	h.DB.Create(&models.AuditLog{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
	})
}

// ---------- Auth ----------

func (h *Handler) Login(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Username != h.Cfg.AdminUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(h.Cfg.AdminPass), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := middleware.GenerateJWT(body.Username, h.Cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// ---------- Dashboard ----------

func (h *Handler) Dashboard(c *gin.Context) {
	now := time.Now()
	soonDate := now.Add(7 * 24 * time.Hour)

	var totalCustomers, activeCustomers, expiringSoon int64
	h.DB.Model(&models.Customer{}).Count(&totalCustomers)
	h.DB.Model(&models.Customer{}).Where("status = 'active'").Count(&activeCustomers)
	h.DB.Model(&models.Customer{}).Where("status = 'active' AND expires_at BETWEEN ? AND ?", now, soonDate).Count(&expiringSoon)

	var onlineServers, offlineServers int64
	h.DB.Model(&models.Server{}).Where("status = 'online'").Count(&onlineServers)
	h.DB.Model(&models.Server{}).Where("status = 'offline'").Count(&offlineServers)

	type revenueRow struct {
		Total float64
	}
	var revenue revenueRow
	h.DB.Model(&models.Customer{}).
		Where("status = 'active' AND STRFTIME('%Y-%m', created_at) = STRFTIME('%Y-%m', 'now')").
		Select("SUM(amount) as total").
		Scan(&revenue)

	var recentLogs []models.AuditLog
	h.DB.Order("created_at desc").Limit(10).Find(&recentLogs)

	c.JSON(http.StatusOK, gin.H{
		"total_customers":  totalCustomers,
		"active_customers": activeCustomers,
		"expiring_soon":    expiringSoon,
		"online_servers":   onlineServers,
		"offline_servers":  offlineServers,
		"monthly_revenue":  revenue.Total,
		"recent_logs":      recentLogs,
	})
}

// ---------- Customers ----------

func (h *Handler) ListCustomers(c *gin.Context) {
	page, perPage := pageParams(c)
	q := h.DB.Model(&models.Customer{})

	if name := c.Query("name"); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if bt := c.Query("billing_type"); bt != "" {
		q = q.Where("billing_type = ?", bt)
	}
	if tag := c.Query("tag"); tag != "" {
		q = q.Where("tags LIKE ?", "%"+tag+"%")
	}
	if ed := c.Query("expiring_days"); ed != "" {
		if days, err := strconv.Atoi(ed); err == nil {
			q = q.Where("expires_at BETWEEN ? AND ?", time.Now(), time.Now().Add(time.Duration(days)*24*time.Hour))
		}
	}

	sortField := c.DefaultQuery("sort", "created_at")
	sortOrder := c.DefaultQuery("order", "desc")
	allowedSort := map[string]bool{"created_at": true, "expires_at": true, "name": true, "amount": true, "status": true}
	if !allowedSort[sortField] {
		sortField = "created_at"
	}
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	q = q.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	var total int64
	q.Count(&total)

	var customers []models.Customer
	q.Offset(offset(page, perPage)).Limit(perPage).Find(&customers)

	c.JSON(http.StatusOK, PageResult{Total: total, Page: page, PerPage: perPage, Data: customers})
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if customer.Status == "" {
		customer.Status = "active"
	}
	if err := h.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.audit("create", "customer", customer.ID, customer.Name, clientIP(c))
	c.JSON(http.StatusCreated, customer)
}

func (h *Handler) GetCustomer(c *gin.Context) {
	var customer models.Customer
	if err := h.DB.First(&customer, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := h.DB.First(&customer, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&customer)
	h.audit("update", "customer", customer.ID, customer.Name, clientIP(c))
	c.JSON(http.StatusOK, customer)
}

func (h *Handler) DeleteCustomer(c *gin.Context) {
	var customer models.Customer
	if err := h.DB.First(&customer, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&customer)
	h.audit("delete", "customer", customer.ID, customer.Name, clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) BatchDeleteCustomers(c *gin.Context) {
	var body struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	h.DB.Delete(&models.Customer{}, body.IDs)
	h.audit("batch_delete", "customer", 0, fmt.Sprintf("ids=%v", body.IDs), clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) BatchRenewCustomers(c *gin.Context) {
	var body struct {
		IDs  []uint `json:"ids"`
		Days int    `json:"days"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 || body.Days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids and days required"})
		return
	}
	var customers []models.Customer
	h.DB.Find(&customers, body.IDs)
	now := time.Now()
	for i := range customers {
		base := customers[i].ExpiresAt
		if base.Before(now) {
			base = now
		}
		customers[i].ExpiresAt = base.Add(time.Duration(body.Days) * 24 * time.Hour)
		customers[i].Status = "active"
		h.DB.Save(&customers[i])
	}
	h.audit("batch_renew", "customer", 0, fmt.Sprintf("ids=%v days=%d", body.IDs, body.Days), clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "renewed", "count": len(customers)})
}

func (h *Handler) ExportCustomers(c *gin.Context) {
	var customers []models.Customer
	h.DB.Find(&customers)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=customers.csv")

	w := csv.NewWriter(c.Writer)
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM for Excel compatibility
	w.Write([]string{"ID", "Name", "Contact", "ExpiresAt", "Amount", "BillingType", "TrafficGB", "UsedGB", "Tags", "Status", "Remark", "CreatedAt"})
	for _, cust := range customers {
		w.Write([]string{
			strconv.Itoa(int(cust.ID)),
			cust.Name,
			cust.Contact,
			cust.ExpiresAt.Format("2006-01-02"),
			strconv.FormatFloat(cust.Amount, 'f', 2, 64),
			cust.BillingType,
			strconv.FormatFloat(cust.TrafficGB, 'f', 2, 64),
			strconv.FormatFloat(cust.UsedGB, 'f', 2, 64),
			cust.Tags,
			cust.Status,
			cust.Remark,
			cust.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
}

// ---------- Regions ----------

func (h *Handler) ListRegions(c *gin.Context) {
	var regions []models.Region
	h.DB.Find(&regions)
	c.JSON(http.StatusOK, regions)
}

func (h *Handler) CreateRegion(c *gin.Context) {
	var region models.Region
	if err := c.ShouldBindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&region)
	h.audit("create", "region", region.ID, region.Name, clientIP(c))
	c.JSON(http.StatusCreated, region)
}

func (h *Handler) UpdateRegion(c *gin.Context) {
	var region models.Region
	if err := h.DB.First(&region, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&region)
	h.audit("update", "region", region.ID, region.Name, clientIP(c))
	c.JSON(http.StatusOK, region)
}

func (h *Handler) DeleteRegion(c *gin.Context) {
	var region models.Region
	if err := h.DB.First(&region, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&region)
	h.audit("delete", "region", region.ID, region.Name, clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---------- Servers ----------

func (h *Handler) ListServers(c *gin.Context) {
	var servers []models.Server
	h.DB.Find(&servers)
	c.JSON(http.StatusOK, servers)
}

func (h *Handler) CreateServer(c *gin.Context) {
	var server models.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if server.Status == "" {
		server.Status = "unknown"
	}
	h.DB.Create(&server)
	h.audit("create", "server", server.ID, server.Name, clientIP(c))
	c.JSON(http.StatusCreated, server)
}

func (h *Handler) UpdateServer(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&server)
	h.audit("update", "server", server.ID, server.Name, clientIP(c))
	c.JSON(http.StatusOK, server)
}

func (h *Handler) DeleteServer(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&server)
	h.audit("delete", "server", server.ID, server.Name, clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) PingServerNow(c *gin.Context) {
	var server models.Server
	if err := h.DB.First(&server, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	latency, _ := services.PingServer(server.IP, 3*time.Second)
	status := "online"
	if latency < 0 {
		status = "offline"
	}
	h.DB.Model(&server).Updates(map[string]interface{}{
		"latency":      latency,
		"status":       status,
		"last_ping_at": time.Now(),
	})
	c.JSON(http.StatusOK, gin.H{"latency": latency, "status": status})
}

// ---------- Routes ----------

func (h *Handler) ListRoutes(c *gin.Context) {
	page, perPage := pageParams(c)
	q := h.DB.Model(&models.Route{}).Preload("Region").Preload("Server")
	if rid := c.Query("region_id"); rid != "" {
		q = q.Where("region_id = ?", rid)
	}
	if sid := c.Query("server_id"); sid != "" {
		q = q.Where("server_id = ?", sid)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var routes []models.Route
	q.Offset(offset(page, perPage)).Limit(perPage).Find(&routes)
	c.JSON(http.StatusOK, PageResult{Total: total, Page: page, PerPage: perPage, Data: routes})
}

func (h *Handler) CreateRoute(c *gin.Context) {
	var route models.Route
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if route.Status == "" {
		route.Status = "active"
	}
	h.DB.Create(&route)
	h.audit("create", "route", route.ID, route.Name, clientIP(c))
	c.JSON(http.StatusCreated, route)
}

func (h *Handler) UpdateRoute(c *gin.Context) {
	var route models.Route
	if err := h.DB.First(&route, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&route)
	h.audit("update", "route", route.ID, route.Name, clientIP(c))
	c.JSON(http.StatusOK, route)
}

func (h *Handler) DeleteRoute(c *gin.Context) {
	var route models.Route
	if err := h.DB.First(&route, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&route)
	h.audit("delete", "route", route.ID, route.Name, clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---------- Nodes ----------

func (h *Handler) ListNodes(c *gin.Context) {
	page, perPage := pageParams(c)
	q := h.DB.Model(&models.Node{}).Preload("Route").Preload("Server")
	if rid := c.Query("route_id"); rid != "" {
		q = q.Where("route_id = ?", rid)
	}
	if sid := c.Query("server_id"); sid != "" {
		q = q.Where("server_id = ?", sid)
	}
	var total int64
	q.Count(&total)
	var nodes []models.Node
	q.Offset(offset(page, perPage)).Limit(perPage).Find(&nodes)
	c.JSON(http.StatusOK, PageResult{Total: total, Page: page, PerPage: perPage, Data: nodes})
}

func (h *Handler) CreateNode(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&node)
	h.audit("create", "node", node.ID, node.Name, clientIP(c))
	c.JSON(http.StatusCreated, node)
}

func (h *Handler) UpdateNode(c *gin.Context) {
	var node models.Node
	if err := h.DB.First(&node, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Save(&node)
	h.audit("update", "node", node.ID, node.Name, clientIP(c))
	c.JSON(http.StatusOK, node)
}

func (h *Handler) DeleteNode(c *gin.Context) {
	var node models.Node
	if err := h.DB.First(&node, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&node)
	h.audit("delete", "node", node.ID, node.Name, clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---------- Audit Logs ----------

func (h *Handler) ListAuditLogs(c *gin.Context) {
	page, perPage := pageParams(c)
	q := h.DB.Model(&models.AuditLog{})
	if action := c.Query("action"); action != "" {
		q = q.Where("action = ?", action)
	}
	if resource := c.Query("resource"); resource != "" {
		q = q.Where("resource = ?", resource)
	}
	var total int64
	q.Count(&total)
	var logs []models.AuditLog
	q.Order("created_at desc").Offset(offset(page, perPage)).Limit(perPage).Find(&logs)
	c.JSON(http.StatusOK, PageResult{Total: total, Page: page, PerPage: perPage, Data: logs})
}

// ---------- API Tokens ----------

func (h *Handler) ListTokens(c *gin.Context) {
	var tokens []models.APIToken
	h.DB.Find(&tokens)
	// Mask tokens in response
	for i := range tokens {
		if len(tokens[i].Token) > 8 {
			tokens[i].Token = tokens[i].Token[:8] + "..."
		}
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) CreateToken(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rawToken, err := generateAPIToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	token := models.APIToken{Name: body.Name, Token: rawToken}
	h.DB.Create(&token)
	h.audit("create", "api_token", token.ID, body.Name, clientIP(c))
	// Return full token once
	c.JSON(http.StatusCreated, token)
}

func (h *Handler) DeleteToken(c *gin.Context) {
	var token models.APIToken
	if err := h.DB.First(&token, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Delete(&token)
	h.audit("delete", "api_token", token.ID, token.Name, clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func generateAPIToken() (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 48)
	randBytes := make([]byte, 48)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	for i := range randBytes {
		b[i] = chars[randBytes[i]%byte(len(chars))]
	}
	return string(b), nil
}

// ---------- Settings ----------

func (h *Handler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"port":           h.Cfg.Port,
		"base_path":      h.Cfg.BasePath,
		"domain":         h.Cfg.Domain,
		"notify_days":    h.Cfg.NotifyDays,
		"ping_interval":  h.Cfg.PingInterval,
		"gmail_from":     h.Cfg.Gmail.FromEmail,
		"gmail_admin":    h.Cfg.Gmail.AdminEmail,
		"gmail_oauth_ok": h.Cfg.Gmail.RefreshToken != "",
	})
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var body struct {
		Domain       string      `json:"domain"`
		NotifyDays   []int       `json:"notify_days"`
		PingInterval int         `json:"ping_interval"`
		Gmail        *config.GmailConfig `json:"gmail"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Domain != "" {
		h.Cfg.Domain = body.Domain
	}
	if len(body.NotifyDays) > 0 {
		h.Cfg.NotifyDays = body.NotifyDays
	}
	if body.PingInterval > 0 {
		h.Cfg.PingInterval = body.PingInterval
	}
	if body.Gmail != nil {
		if body.Gmail.ClientID != "" {
			h.Cfg.Gmail.ClientID = body.Gmail.ClientID
		}
		if body.Gmail.ClientSecret != "" {
			h.Cfg.Gmail.ClientSecret = body.Gmail.ClientSecret
		}
		if body.Gmail.FromEmail != "" {
			h.Cfg.Gmail.FromEmail = body.Gmail.FromEmail
		}
		if body.Gmail.AdminEmail != "" {
			h.Cfg.Gmail.AdminEmail = body.Gmail.AdminEmail
		}
	}
	if err := config.Save(h.CfgPath, h.Cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	h.audit("update", "settings", 0, "settings updated", clientIP(c))
	c.JSON(http.StatusOK, gin.H{"message": "saved"})
}

func (h *Handler) GmailAuthURL(c *gin.Context) {
	if h.Cfg.Gmail.ClientID == "" || h.Cfg.Gmail.ClientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gmail client_id and client_secret must be configured first"})
		return
	}
	authURL := services.GetOAuthURL(h.Cfg.Gmail.ClientID, h.Cfg.Gmail.ClientSecret)
	c.JSON(http.StatusOK, gin.H{"url": authURL})
}

func (h *Handler) GmailCallback(c *gin.Context) {
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refreshToken, email, err := services.ExchangeCode(h.Cfg.Gmail.ClientID, h.Cfg.Gmail.ClientSecret, body.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.Cfg.Gmail.RefreshToken = refreshToken
	if email != "" && h.Cfg.Gmail.FromEmail == "" {
		h.Cfg.Gmail.FromEmail = email
	}
	if err := config.Save(h.CfgPath, h.Cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "gmail configured", "email": email})
}

// ---------- System ----------

func (h *Handler) SystemVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": Version})
}

func (h *Handler) SystemUpdate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "update triggered (not implemented in this build)"})
}

// Version is set at build time via ldflags.
var Version = "dev"
