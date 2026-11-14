package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
)

// Keep protocol revisions used by previous-minor clients negotiable after SDK upgrades.
func TestPreviousClientProtocolInitialization(t *testing.T) {
	for _, version := range []string{"2024-11-05", "2025-03-26", "2025-06-18"} {
		t.Run(version, func(t *testing.T) {
			r, token, cleanup := setupAuthTest(t, models.Enabled, 1)
			defer cleanup()
			body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"` + version + `","capabilities":{},"clientInfo":{"name":"previous-client","version":"1.6.1"}}}`
			req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("initialize: status=%d body=%s", w.Code, w.Body.String())
			}
			payload := w.Body.String()
			if strings.Contains(w.Header().Get("Content-Type"), "text/event-stream") {
				for _, line := range strings.Split(payload, "\n") {
					if strings.HasPrefix(line, "data: ") {
						payload = strings.TrimPrefix(line, "data: ")
						break
					}
				}
			}
			var response struct {
				Result struct {
					ProtocolVersion string `json:"protocolVersion"`
				} `json:"result"`
			}
			if err := json.Unmarshal([]byte(payload), &response); err != nil {
				t.Fatal(err)
			}
			if response.Result.ProtocolVersion != version {
				t.Fatalf("protocol=%q want %q", response.Result.ProtocolVersion, version)
			}
		})
	}
}
