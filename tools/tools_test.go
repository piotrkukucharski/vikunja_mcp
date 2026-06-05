package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"vikunja_mcp/client"
)

func TestHandleEntityAllEntities(t *testing.T) {
	entities := []string{
		"projects",
		"tasks",
		"filters",
		"labels",
		"projects_views",
		"projects_views_tasks",
		"reactions",
		"tasks_assignees",
		"tasks_attachments",
		"tasks_comments",
		"tasks_labels",
		"tasks_relations",
	}

	actions := []string{"create", "read", "update", "delete"}

	for _, entity := range entities {
		for _, action := range actions {
			t.Run(entity+"_"+action, func(t *testing.T) {
				expectedMethod := ""
				switch action {
				case "create":
					expectedMethod = "PUT"
				case "read":
					expectedMethod = "GET"
				case "update":
					expectedMethod = "POST"
				case "delete":
					expectedMethod = "DELETE"
				}

				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != expectedMethod {
						t.Errorf("Expected method %s, got %s", expectedMethod, r.Method)
					}
					if r.Header.Get("Authorization") != "Bearer test-token" {
						t.Errorf("Expected token Bearer test-token, got %s", r.Header.Get("Authorization"))
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status": "ok"}`))
				}))
				defer ts.Close()

				c := client.NewVikunjaClient(ts.URL, "Bearer test-token")
				input := EntityInput{
					Action: action,
					Path:   "/api/v1/" + entity,
					Data:   map[string]any{"test": "data"},
				}

				res, err := handleEntity(context.Background(), c, input)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if res.IsError {
					t.Fatalf("Tool returned error: %+v", res.Content)
				}
			})
		}
	}
}

func TestHandleEntityInvalidAction(t *testing.T) {
	c := client.NewVikunjaClient("http://localhost:8080", "Bearer test")
	input := EntityInput{
		Action: "invalid",
		Path:   "/api/v1/projects",
	}
	_, err := handleEntity(context.Background(), c, input)
	if err == nil {
		t.Fatal("Expected error for invalid action, got nil")
	}
}
