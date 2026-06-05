# Vikunja MCP Server

A Model Context Protocol (MCP) server for [Vikunja](https://vikunja.io/), built with Go.

This server enables AI models and clients supporting the MCP standard to interact with your Vikunja instance using HTTP Server-Sent Events (SSE). It supports essential CRUD operations across multiple Vikunja entities.

## Features

- **HTTP/SSE Transport**: Runs over HTTP using the `go-sdk`'s `SSEServerTransport`.
- **Pass-through Authentication**: Extracts the `Authorization` header from incoming MCP connections and forwards it securely to the Vikunja API.
- **Generic CRUD Support**: Includes customizable MCP tools for managing the following entities:
  - Projects
  - Tasks
  - Filters
  - Labels
  - Projects Views
  - Projects Views Tasks
  - Reactions
  - Tasks Assignees
  - Tasks Attachments
  - Tasks Comments
  - Tasks Labels
  - Tasks Relations
- **Container Ready**: Includes a multi-stage `Dockerfile` and GitHub Actions CI workflow for automatic testing, building, and publishing.

## Prerequisites

- Go 1.26+ (if building from source)
- Docker (if running via container)
- A running Vikunja instance

## Configuration

The server requires the base URL of your Vikunja API. This is provided via environment variables.

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `VIKUNJA_BASE_URL`   | **Required.** The root URL for the Vikunja API (e.g., `https://api.yourvikunja.com`). | - |
| `PORT`               | The port the MCP server will listen on. | `8080` |

## Usage

### Running Locally

```bash
# Clone the repository
git clone https://github.com/user/vikunja_mcp.git
cd vikunja_mcp

# Export required environment variables
export VIKUNJA_BASE_URL="https://api.vikunja.example.com"
export PORT="8080"

# Run the server
go run main.go
```

### Running with Docker

```bash
docker run -p 8080:8080 -e VIKUNJA_BASE_URL="https://api.vikunja.example.com" ghcr.io/yourusername/vikunja-mcp:latest
```

### Running with Docker Compose

You can deploy the MCP server using Docker Compose. Here is an example config:

```yaml
version: '3.8'

services:
  vikunja-mcp:
    image: ghcr.io/yourusername/vikunja-mcp:latest
    ports:
      - "8080:8080"
    environment:
      - VIKUNJA_BASE_URL=https://api.vikunja.example.com # Or http://vikunja:3456 if on the same network
      - PORT=8080
    restart: unless-stopped
```

Run it with:
```bash
docker compose up -d
```

## How to Connect (MCP Client)

Connect your MCP client to the server's SSE endpoint (in this case, it mounts on the root `http://localhost:8080/`).
Make sure your client is configured to send the required Vikunja JWT token as a Bearer token in the request headers:

```http
Authorization: Bearer <your-vikunja-token>
```

## Available Tools

The MCP server exposes a `manage_{entity}` tool for each supported resource. For example, `manage_tasks` or `manage_projects`.

**Tool Schema (Generic):**
- `action` (string, required): The CRUD action to perform (`create`, `read`, `update`, `delete`).
- `path` (string, required): The specific API path suffix (e.g., `/api/v1/projects` or `/api/v1/projects/1`).
- `data` (object, optional): The JSON payload used for `create` and `update` actions.
