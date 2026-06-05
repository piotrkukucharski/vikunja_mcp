package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"vikunja_mcp/client"
	"vikunja_mcp/tools"
)

func makeHealthHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Check reachability by calling public API info endpoint
		tempClient := client.NewVikunjaClient(baseURL, "")
		var info any
		err := tempClient.Request(r.Context(), "GET", "/api/v1/info", nil, &info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"status":  "error",
				"error":   "Vikunja service is unreachable or not working",
				"details": err.Error(),
			})
			return
		}

		authHeader := r.Header.Get("Authorization")
		authChecked := false
		if authHeader != "" {
			authChecked = true
			authClient := client.NewVikunjaClient(baseURL, authHeader)
			var tasks any
			err := authClient.Request(r.Context(), "GET", "/api/v1/tasks?limit=1", nil, &tasks)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				errStr := err.Error()
				errType := "Failed to fetch tasks"
				if strings.Contains(strings.ToLower(errStr), "401") || strings.Contains(strings.ToLower(errStr), "unauthorized") ||
					strings.Contains(strings.ToLower(errStr), "403") || strings.Contains(strings.ToLower(errStr), "forbidden") {
					errType = "Vikunja API token is invalid or unauthorized"
				}
				json.NewEncoder(w).Encode(map[string]any{
					"status":  "error",
					"error":   errType,
					"details": errStr,
				})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"status":            "ok",
			"vikunja_reachable": true,
			"auth_checked":      authChecked,
		})
	}
}

func main() {
	baseURL := os.Getenv("VIKUNJA_BASE_URL")
	if baseURL == "" {
		log.Fatal("VIKUNJA_BASE_URL environment variable is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sse := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
		token := req.Header.Get("Authorization")
		if token == "" {
			// If no token is provided, the request will fail or we can reject it.
			// Currently, we just pass the empty token and let the API fail if it needs auth.
		}

		c := client.NewVikunjaClient(baseURL, token)
		server := mcp.NewServer(&mcp.Implementation{
			Name:    "vikunja_mcp",
			Version: "1.0.0",
		}, nil)

		tools.RegisterAll(server, c)

		return server
	}, &mcp.SSEOptions{
		// Optional SSE configuration can go here
	})

	http.HandleFunc("/health", makeHealthHandler(baseURL))

	http.Handle("/", sse)

	log.Printf("Starting Vikunja MCP server on port %s", port)
	log.Printf("Using Vikunja Base URL: %s", baseURL)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
