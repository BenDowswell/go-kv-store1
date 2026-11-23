package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BenDowswell/kvstore1/store"
)

func newRouter() http.Handler {
	return NewHandler(store.NewKVService(store.NewMemoryStore())).Routes()
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	if rec := do(newRouter(), "GET", "/health", ""); rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestGetMissing(t *testing.T) {
	if rec := do(newRouter(), "GET", "/kv/nope", ""); rec.Code != 404 {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestPut(t *testing.T) {
	tests := []struct {
		name, path, body string
		want             int
	}{
		{"valid", "/kv/a", `{"value":"1"}`, 204},
		{"bad json", "/kv/a", `{`, 400},
		{"empty value", "/kv/a", `{"value":""}`, 400},
		{"blank key", "/kv/%20", `{"value":"1"}`, 400},
		{"too large", "/kv/a", `{"value":"` + strings.Repeat("x", maxBodySize) + `"}`, 413},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(newRouter(), "PUT", tt.path, tt.body); rec.Code != tt.want {
				t.Fatalf("code %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestDeleteBlankKey(t *testing.T) {
	if rec := do(newRouter(), "DELETE", "/kv/%20", ""); rec.Code != 400 {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	if rec := do(newRouter(), "POST", "/kv/a", ""); rec.Code != 405 {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestRoundTrip(t *testing.T) {
	h := newRouter()
	if rec := do(h, "PUT", "/kv/a", `{"value":"hello"}`); rec.Code != 204 {
		t.Fatalf("put %d", rec.Code)
	}
	rec := do(h, "GET", "/kv/a", "")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("get %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	var env struct {
		Data ResponsePayload `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Key != "a" || env.Data.Value != "hello" {
		t.Fatalf("got %+v", env.Data)
	}
	if rec := do(h, "DELETE", "/kv/a", ""); rec.Code != 204 {
		t.Fatalf("delete %d", rec.Code)
	}
	if rec := do(h, "GET", "/kv/a", ""); rec.Code != 404 {
		t.Fatalf("get after delete %d", rec.Code)
	}
}

func TestStartServerShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- StartServer(ctx, "127.0.0.1:0", store.NewKVService(store.NewMemoryStore()))
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestStartServerListenError(t *testing.T) {
	err := StartServer(context.Background(), "bad-address", store.NewKVService(store.NewMemoryStore()))
	if err == nil {
		t.Fatal("expected error")
	}
}
