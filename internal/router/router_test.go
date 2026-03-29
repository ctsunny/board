package router

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/ctsunny/board/internal/api"
	"github.com/ctsunny/board/internal/config"
)

func TestBasePathRedirectsToTrailingSlash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	staticFiles := fstest.MapFS{
		"index.html":      {Data: []byte("<html><head></head><body>ok</body></html>")},
		"assets/app.js":   {Data: []byte("console.log('ok')")},
		"assets/app.css":  {Data: []byte("body{}")},
	}

	Setup(engine, &api.Handler{}, &config.Config{BasePath: "/mgmt-test"}, nil, fs.FS(staticFiles))

	req := httptest.NewRequest(http.MethodGet, "/mgmt-test", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusPermanentRedirect {
		t.Fatalf("expected %d, got %d", http.StatusPermanentRedirect, rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/mgmt-test/" {
		t.Fatalf("expected redirect to /mgmt-test/, got %q", got)
	}
}

func TestStaticAssetsAndIndexUseBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	staticFiles := fstest.MapFS{
		"index.html":    {Data: []byte("<html><head></head><body>ok</body></html>")},
		"assets/app.js": {Data: []byte("console.log('ok')")},
	}

	Setup(engine, &api.Handler{}, &config.Config{BasePath: "/mgmt-test"}, nil, fs.FS(staticFiles))

	indexReq := httptest.NewRequest(http.MethodGet, "/mgmt-test/", nil)
	indexRec := httptest.NewRecorder()
	engine.ServeHTTP(indexRec, indexReq)

	if indexRec.Code != http.StatusOK {
		t.Fatalf("expected %d for index, got %d", http.StatusOK, indexRec.Code)
	}
	if body := indexRec.Body.String(); !strings.Contains(body, `window.__BOARD_BASE__="/mgmt-test"`) {
		t.Fatalf("expected injected base path in index response, got %q", body)
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/mgmt-test/assets/app.js", nil)
	assetRec := httptest.NewRecorder()
	engine.ServeHTTP(assetRec, assetReq)

	if assetRec.Code != http.StatusOK {
		t.Fatalf("expected %d for asset, got %d", http.StatusOK, assetRec.Code)
	}
	if body := assetRec.Body.String(); !strings.Contains(body, "console.log('ok')") {
		t.Fatalf("expected asset body to be served, got %q", body)
	}
}
