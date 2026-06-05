package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"vikunja_mcp/client"
)

type EntityInput struct {
	Action string         `json:"action" jsonschema:"Action: create, read, update, delete"`
	Path   string         `json:"path" jsonschema:"API Path suffix, e.g. /api/v1/projects or /api/v1/projects/1. Must match the entity type."`
	Data   map[string]any `json:"data,omitempty" jsonschema:"JSON data for create/update"`
}

func formatResult(data any, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
			},
			IsError: true,
		}, nil
	}
	if data == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Success"},
			},
		}, nil
	}
	b, _ := json.MarshalIndent(data, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(b)},
		},
	}, nil
}

func handleEntity(ctx context.Context, c *client.VikunjaClient, input EntityInput) (*mcp.CallToolResult, error) {
	if !strings.HasPrefix(input.Path, "/") {
		input.Path = "/" + input.Path
	}

	switch input.Action {
	case "create":
		var result map[string]any
		err := c.Request(ctx, "PUT", input.Path, input.Data, &result)
		return formatResult(result, err)
	case "read":
		var result any
		err := c.Request(ctx, "GET", input.Path, nil, &result)
		return formatResult(result, err)
	case "update":
		var result map[string]any
		err := c.Request(ctx, "POST", input.Path, input.Data, &result)
		return formatResult(result, err)
	case "delete":
		err := c.Request(ctx, "DELETE", input.Path, nil, nil)
		return formatResult(nil, err)
	default:
		return nil, fmt.Errorf("unknown action: %s", input.Action)
	}
}

func RegisterAll(server *mcp.Server, c *client.VikunjaClient) {
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

	for _, entity := range entities {
		e := entity // capture loop variable
		mcp.AddTool(server, &mcp.Tool{
			Name:        fmt.Sprintf("manage_%s", e),
			Description: fmt.Sprintf("CRUD operations for %s", strings.ReplaceAll(e, "_", " ")),
		}, func(ctx context.Context, req *mcp.CallToolRequest, input EntityInput) (*mcp.CallToolResult, any, error) {
			res, err := handleEntity(ctx, c, input)
			return res, nil, err
		})
	}
}
