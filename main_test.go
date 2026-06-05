package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HealthResponse struct {
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
	Details          string `json:"details,omitempty"`
	VikunjaReachable bool   `json:"vikunja_reachable,omitempty"`
	AuthChecked      bool   `json:"auth_checked,omitempty"`
}

func TestHealthEndpointScenarios(t *testing.T) {
	// 1. Scenario: Vikunja service is down / unreachable
	t.Run("VikunjaUnreachable", func(t *testing.T) {
		handler := makeHealthHandler("http://localhost:54321") // Invalid port to simulate downtime

		req := httptest.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		handler(wWriter(rr), req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", rr.Code)
		}

		var resp HealthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if resp.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", resp.Status)
		}
		if resp.Error != "Vikunja service is unreachable or not working" {
			t.Errorf("Expected reachability error, got '%s'", resp.Error)
		}
	})

	// Setup mock Vikunja Server for reachable cases
	var mockVikunjaResponse func(w http.ResponseWriter, r *http.Request)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockVikunjaResponse(w, r)
	}))
	defer ts.Close()

	// 2. Scenario: Reachable, no auth token provided
	t.Run("ReachableNoAuth", func(t *testing.T) {
		mockVikunjaResponse = func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/info" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"version": "1.0.0"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}

		handler := makeHealthHandler(ts.URL)
		req := httptest.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()

		handler(wWriter(rr), req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var resp HealthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if resp.Status != "ok" || !resp.VikunjaReachable || resp.AuthChecked {
			t.Errorf("Unexpected response: %+v", resp)
		}
	})

	// 3. Scenario: Reachable, auth token provided but invalid (401)
	t.Run("ReachableInvalidAuth", func(t *testing.T) {
		mockVikunjaResponse = func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/info" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"version": "1.0.0"}`))
				return
			}
			if r.URL.Path == "/api/v1/tasks" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "unauthorized"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}

		handler := makeHealthHandler(ts.URL)
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler(wWriter(rr), req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", rr.Code)
		}

		var resp HealthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if resp.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", resp.Status)
		}
		if resp.Error != "Vikunja API token is invalid or unauthorized" {
			t.Errorf("Expected token invalid error, got '%s'", resp.Error)
		}
	})

	// 4. Scenario: Reachable, auth token provided and valid
	t.Run("ReachableValidAuth", func(t *testing.T) {
		mockVikunjaResponse = func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/info" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"version": "1.0.0"}`))
				return
			}
			if r.URL.Path == "/api/v1/tasks" {
				if r.Header.Get("Authorization") == "Bearer valid-token" && r.URL.Query().Get("limit") == "1" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[{"id": 1, "title": "test"}]`))
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}

		handler := makeHealthHandler(ts.URL)
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler(wWriter(rr), req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var resp HealthResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if resp.Status != "ok" || !resp.VikunjaReachable || !resp.AuthChecked {
			t.Errorf("Unexpected response: %+v", resp)
		}
	})
}

// Helper wrapper because Go httptest.ResponseRecorder does not require specific conversions
// but sometimes compiler prefers direct type or casting.
func wWriter(w *httptest.ResponseRecorder) http.ResponseWriter {
	return w
}
