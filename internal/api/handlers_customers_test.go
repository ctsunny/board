package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/ctsunny/board/internal/config"
	dbpkg "github.com/ctsunny/board/internal/db"
	"github.com/ctsunny/board/internal/models"
	"github.com/gin-gonic/gin"
)

func TestListCustomersSupportsRegionTagFilterAndSorting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	database, err := dbpkg.Init(filepath.Join(tmpDir, "board.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	customers := []models.Customer{
		{
			Name:       "A",
			Contact:    "13800000001",
			RegionName: "美国, 英国",
			Amount:     50,
			Status:     "active",
			ExpiresAt:  time.Now().Add(48 * time.Hour),
		},
		{
			Name:       "B",
			Contact:    "13800000002",
			RegionName: "日本",
			Amount:     10,
			Status:     "active",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		},
	}
	if err := database.Create(&customers).Error; err != nil {
		t.Fatalf("seed customers: %v", err)
	}

	h := &Handler{
		DB:  database,
		Cfg: &config.Config{},
	}

	filterRec := httptest.NewRecorder()
	filterCtx, _ := gin.CreateTestContext(filterRec)
	filterCtx.Request = httptest.NewRequest(http.MethodGet, "/customers?tag=英国", nil)
	h.ListCustomers(filterCtx)

	if filterRec.Code != http.StatusOK {
		t.Fatalf("expected filter status %d, got %d", http.StatusOK, filterRec.Code)
	}

	var filterBody struct {
		Total int64                    `json:"total"`
		Data  []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(filterRec.Body.Bytes(), &filterBody); err != nil {
		t.Fatalf("unmarshal filter body: %v", err)
	}
	if filterBody.Total != 1 {
		t.Fatalf("expected 1 filtered customer, got %d", filterBody.Total)
	}
	if got := filterBody.Data[0]["name"]; got != "A" {
		t.Fatalf("expected filtered customer A, got %#v", got)
	}

	sortRec := httptest.NewRecorder()
	sortCtx, _ := gin.CreateTestContext(sortRec)
	sortCtx.Request = httptest.NewRequest(http.MethodGet, "/customers?page=1&page_size=1&sort=amount&order=asc", nil)
	h.ListCustomers(sortCtx)

	if sortRec.Code != http.StatusOK {
		t.Fatalf("expected sort status %d, got %d", http.StatusOK, sortRec.Code)
	}

	var sortBody struct {
		PerPage int                      `json:"per_page"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(sortRec.Body.Bytes(), &sortBody); err != nil {
		t.Fatalf("unmarshal sort body: %v", err)
	}
	if sortBody.PerPage != 1 {
		t.Fatalf("expected page_size fallback to set per_page=1, got %d", sortBody.PerPage)
	}
	if got := sortBody.Data[0]["name"]; got != "B" {
		t.Fatalf("expected lowest amount customer B first, got %#v", got)
	}
}
