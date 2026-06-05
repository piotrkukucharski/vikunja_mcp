package main

import (
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"vikunja_mcp/client"
	"vikunja_mcp/tools"
)

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

	http.Handle("/", sse)

	log.Printf("Starting Vikunja MCP server on port %s", port)
	log.Printf("Using Vikunja Base URL: %s", baseURL)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
