package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestDebugRedirect_IndexHTML(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// Test 1: fstest.MapFS + FileFromFS
	rootFS := fstest.MapFS{
		"index.html": {Data: []byte("<html></html>")},
	}
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(rootFS))
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	t.Logf("MapFS+FileFromFS: Status=%d, Loc=%v, Body=%q", w.Code, w.Header()["Location"], w.Body.String())

	// Test 2: http.Dir + FileFromFS (real temp dir)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0644)
	r2 := gin.New()
	r2.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", http.Dir(dir))
	})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", nil)
	r2.ServeHTTP(w2, req2)
	t.Logf("Dir+FileFromFS: Status=%d, Loc=%v, Body=%q", w2.Code, w2.Header()["Location"], w2.Body.String())
}
