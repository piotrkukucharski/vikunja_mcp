package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"vikunja_mcp/client"
)

func TestIntegrationVikunja(t *testing.T) {
	ctx := context.Background()

	// 1. Skip if docker is not available
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		t.Skipf("Docker is not available, skipping integration test: %v", err)
	}
	defer provider.Close()

	// 2. Start Vikunja Container
	t.Log("Starting Vikunja container...")
	req := testcontainers.ContainerRequest{
		Image:        "vikunja/vikunja:latest",
		ExposedPorts: []string{"3456/tcp"},
		Env: map[string]string{
			"VIKUNJA_SERVICE_SECRET": "super-secret-random-key-for-jwt-signing-that-is-long-enough",
			"VIKUNJA_CORS_ENABLE":    "false",
			"VIKUNJA_FILES_BASEPATH": "/tmp/files",
			"VIKUNJA_DATABASE_PATH":  "/tmp/vikunja.db",
		},
		WaitingFor: wait.ForListeningPort("3456/tcp").WithStartupTimeout(60 * time.Second),
	}

	vikunjaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Vikunja container: %v", err)
	}
	defer func() {
		if err := vikunjaC.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate container: %v", err)
		}
	}()

	// 3. Get connection details
	ip, err := vikunjaC.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}
	port, err := vikunjaC.MappedPort(ctx, "3456")
	if err != nil {
		t.Fatalf("Failed to get container mapped port: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s:%s", ip, port.Port())
	t.Logf("Vikunja API is running at: %s", baseURL)

	// Wait a moment for database migrations to settle inside the container
	time.Sleep(3 * time.Second)

	// 4. Register a test user
	t.Log("Registering test user...")
	registerPayload := `{"username": "testuser", "email": "test@example.com", "password": "super-password-123"}`
	resp, err := http.Post(baseURL+"/api/v1/register", "application/json", strings.NewReader(registerPayload))
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to register user. Status: %d", resp.StatusCode)
	}

	// 5. Login to get token
	t.Log("Logging in...")
	loginPayload := `{"username": "testuser", "password": "super-password-123"}`
	respLogin, err := http.Post(baseURL+"/api/v1/login", "application/json", strings.NewReader(loginPayload))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer respLogin.Body.Close()
	if respLogin.StatusCode != http.StatusOK {
		t.Fatalf("Failed to login. Status: %d", respLogin.StatusCode)
	}

	var loginResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(respLogin.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if loginResponse.Token == "" {
		t.Fatal("Login response token is empty")
	}
	tokenHeader := "Bearer " + loginResponse.Token

	// 6. Test MCP Tools via VikunjaClient
	t.Log("Testing MCP client and tools against live container...")
	c := client.NewVikunjaClient(baseURL, tokenHeader)

	// Create project
	createInput := EntityInput{
		Action: "create",
		Path:   "/api/v1/projects",
		Data: map[string]any{
			"title": "Integration Test Project",
		},
	}

	res, err := handleEntity(ctx, c, createInput)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if res.IsError {
		t.Fatalf("Create project tool returned error: %s", res.Content[0].(*mcp.TextContent).Text)
	}

	// Extract created project ID to read it
	var project struct {
		ID int `json:"id"`
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &project); err != nil {
		t.Fatalf("Failed to parse project response: %v, text was: %s", err, text)
	}

	if project.ID == 0 {
		t.Fatalf("Created project ID is 0")
	}

	// Read the project
	readInput := EntityInput{
		Action: "read",
		Path:   fmt.Sprintf("/api/v1/projects/%d", project.ID),
	}
	resRead, err := handleEntity(ctx, c, readInput)
	if err != nil {
		t.Fatalf("Failed to read project: %v", err)
	}
	if resRead.IsError {
		t.Fatalf("Read project tool returned error: %s", resRead.Content[0].(*mcp.TextContent).Text)
	}

	var readProject struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}
	textRead := resRead.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(textRead), &readProject); err != nil {
		t.Fatalf("Failed to parse read project response: %v, text was: %s", err, textRead)
	}

	if readProject.ID != project.ID {
		t.Errorf("Expected project ID %d, got %d", project.ID, readProject.ID)
	}
	if readProject.Title != "Integration Test Project" {
		t.Errorf("Expected project title 'Integration Test Project', got '%s'", readProject.Title)
	}
}
