package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVikunjaClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected token 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}
		if r.URL.Path == "/api/v1/projects" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 1, "title": "Test"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewVikunjaClient(ts.URL, "Bearer test-token")
	var result map[string]any
	err := client.Request(context.Background(), "GET", "/api/v1/projects", nil, &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["id"] != float64(1) {
		t.Errorf("Expected id 1, got %v", result["id"])
	}
}
