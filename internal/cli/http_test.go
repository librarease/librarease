package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIClientDoEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"ok": true},
			"meta": map[string]any{"total": 1, "skip": 0, "limit": 1},
		})
	}))
	defer srv.Close()

	cfg := &Config{
		BaseURL: srv.URL,
		Token:   "test-token",
		Timeout: 5 * time.Second,
		Output:  OutputJSON,
	}
	api := NewAPIClient(cfg)
	resp, data, err := api.Do(http.MethodGet, "/api/v1/users", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	env, err := decodeEnvelope(data)
	if err != nil {
		t.Fatal(err)
	}
	if env.Meta == nil || env.Meta.Total != 1 {
		t.Fatalf("unexpected meta: %+v", env.Meta)
	}
}

func TestHTTPError(t *testing.T) {
	resp := &http.Response{StatusCode: 400}
	err := httpError(resp, []byte(`{"error":"bad request"}`))
	if err == nil || err.Error() == "" {
		t.Fatal("expected error")
	}
}

