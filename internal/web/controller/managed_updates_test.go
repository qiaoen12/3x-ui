package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpsManagedUpdatesRoutesRejectAllMutations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	(&ServerController{}).initRouter(e.Group("/panel/api/server"))
	(&NodeController{}).initRouter(e.Group("/panel/api/nodes"))
	for _, path := range []string{
		"/panel/api/server/installXray/v26.10.10",
		"/panel/api/server/updatePanel",
		"/panel/api/server/setUpdateChannel",
		"/panel/api/server/updateGeofile",
		"/panel/api/server/updateGeofile/geoip.dat",
		"/panel/api/nodes/updatePanel",
	} {
		for _, body := range []string{"", "dev=true", "dev=invalid", `{"ids":[1],"dev":true}`} {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"success":false`) || !strings.Contains(w.Body.String(), "pinned operations") {
				t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
			}
		}
	}
	// These controls remain routed to their original handlers; only updates are denied.
	for _, r := range e.Routes() {
		if strings.HasSuffix(r.Path, "/restartXrayService") || strings.HasSuffix(r.Path, "/stopXrayService") || strings.HasSuffix(r.Path, "/status") {
			if strings.Contains(r.Handler, "rejectBuiltinUpdate") {
				t.Fatalf("ordinary control disabled: %s", r.Path)
			}
		}
	}
}
