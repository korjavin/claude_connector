# End-to-End Implementation and Deployment Guide

## 2.1. Prerequisites and Environment Setup

Before beginning the implementation, ensure the following software and access credentials are in place. This guide assumes a development environment running a Unix-like operating system (Linux or macOS).

- **Go**: The latest stable version of the Go programming language. Verify the installation by running `go version` in a terminal.
- **Docker & Docker Compose**: The Docker Engine and Docker Compose tool are required for containerization and orchestration. Verify installations with `docker --version` and `docker compose version`.
- **Code Editor**: A modern code editor such as Visual Studio Code with the official Go extension is highly recommended for features like IntelliSense, formatting, and debugging.
- **Server Access**: SSH access to the server where the application will be deployed. This server must have Docker and Portainer installed.
- **Portainer Access**: A username and password for the Portainer management UI running on the server.
- **Cloudflare Account**: Access to a Cloudflare account that manages the DNS for a registered domain name.

## 2.2. Go Project Initialization

This section covers the setup of the Go project structure and the installation of necessary third-party libraries.

First, create the project directory structure. The data subdirectory will hold the sensitive CSV file, ensuring it is kept separate from the application source code.

```bash
mkdir -p claude-connector/data
cd claude-connector
```

Next, initialize a new Go module. This command creates a go.mod file, which tracks the project's dependencies. Replace github.com/user/claude-connector with your own module path, typically your Git repository URL.

```bash
go mod init github.com/user/claude-connector
```

Now, install the required Go packages using go get. This project will leverage two key libraries:

- **Gin Web Framework** (github.com/gin-gonic/gin): A high-performance, minimalist web framework for Go. Its popularity, extensive documentation, and simple API make it an excellent choice for building the HTTP server that will receive requests from Claude.
- **mcp-golang** (github.com/metoro-io/mcp-golang): A third-party Go library that provides a high-level, type-safe abstraction over the Model Context Protocol. It simplifies the process of defining tools and handling the underlying JSON-RPC 2.0 communication, allowing the developer to focus on the tool's business logic rather than protocol boilerplate.

Execute the following command to download and install these dependencies:

```bash
go get github.com/gin-gonic/gin github.com/metoro-io/mcp-golang
```

After running this, the go.mod file will be updated with the new dependencies, and a go.sum file will be created to record the checksums of these dependencies, ensuring reproducible builds.

## 2.3. Implementing the Go MCP Server

The core of the project is the Go application that acts as the MCP server. The code will be organized into logical packages for clarity and maintainability.

### 2.3.1. Main Application Entrypoint (main.go)

Create a file named main.go in the root of the project directory. This file will be responsible for loading configuration, setting up the Gin router, applying middleware, defining the API endpoint, and starting the server.

```go
// main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/user/claude-connector/handlers"
	"github.com/user/claude-connector/middleware"
)

func main() {
	// Best practice: load configuration from environment variables
	port := os.Getenv("MCP_SERVER_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	apiKey := os.Getenv("API_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: API_SECRET_KEY environment variable not set.")
	}

	csvPath := os.Getenv("CSV_FILE_PATH")
	if csvPath == "" {
		log.Fatal("FATAL: CSV_FILE_PATH environment variable not set.")
	}

	// Initialize Gin router
	router := gin.Default()

	// MCP specifies a single endpoint, typically /mcp, that handles all methods
	mcpGroup := router.Group("/mcp")
	{
		// Apply the authentication middleware to this group
		mcpGroup.Use(middleware.AuthMiddleware(apiKey))
		// Handle all POST requests to /mcp
		mcpGroup.POST("", handlers.MCPHandler(csvPath))
	}

	// Start the server
	log.Printf("Starting MCP server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

### 2.3.2. Authentication Middleware (middleware/auth.go)

Create a new directory middleware and inside it, a file named auth.go. This middleware will protect the /mcp endpoint by validating a Bearer token on every incoming request.

```go
// middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a Gin middleware for API key authentication.
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Use 'Bearer <token>'"})
			return
		}

		if parts[1] != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			return
		}

		c.Next()
	}
}
```

### 2.3.3. MCP Handler (handlers/mcp_handler.go)

Create a directory handlers and a file mcp_handler.go inside it. This is where the mcp-golang library is used to define the tool and process MCP requests.

```go
// handlers/mcp_handler.go
package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
	"github.com/user/claude-connector/tools"
)

// Define the structure for the tool's arguments.
// The `json` and `jsonschema` tags are used by mcp-golang to
// automatically generate the schema for Claude.
type GetLastNRecordsArgs struct {
	Count int `json:"count" jsonschema:"required,description=The number of recent records to retrieve."`
}

// MCPHandler initializes the MCP server and returns a Gin handler.
func MCPHandler(csvPath string) gin.HandlerFunc {
	// Use the Gin transport provided by mcp-golang
	transport := http.NewGinTransport()
	server := mcp.NewServer(transport)

	// Register the tool
	err := server.RegisterTool(
		"get_last_n_records",
		"Retrieves the last N records from the local medical information CSV file.",
		func(args GetLastNRecordsArgs) (*mcp.ToolResponse, error) {
			if args.Count <= 0 {
				return nil, fmt.Errorf("count must be a positive integer")
			}

			records, err := tools.GetLastNRecords(csvPath, args.Count)
			if err != nil {
				return nil, fmt.Errorf("failed to get records: %w", err)
			}

			// The tool result must be returned as a text content block.
			// We format the records back into a CSV string.
			var result string
			for _, record := range records {
				result += strings.Join(record, ",") + "\n"
			}

			return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
		},
	)

	if err != nil {
		panic(fmt.Sprintf("Failed to register tool: %v", err))
	}

	return transport.Handler()
}
```

### 2.3.4. Tool Logic (tools/csv_reader.go)

Finally, create a tools directory and a csv_reader.go file. This module contains the core business logic for reading and processing the CSV file.

The standard Go encoding/csv library does not provide a direct method for reading the last N lines of a file efficiently without reading the entire file. Given that personal medical data files are typically manageable in size (i.e., not terabytes), the most robust and straightforward approach is to read the entire file into memory and then slice the resulting data structure. This prioritizes correctness and implementation simplicity over memory optimization, which is an acceptable trade-off for this use case.

```go
// tools/csv_reader.go
package tools

import (
	"encoding/csv"
	"fmt"
	"os"
)

// GetLastNRecords reads a CSV file and returns the last n records.
func GetLastNRecords(filePath string, n int) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open csv file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read csv file: %w", err)
	}

	totalRecords := len(records)
	if totalRecords == 0 {
		return [][]string{}, nil // Return empty slice if file is empty
	}

	// Determine the starting index for the slice
	startIndex := totalRecords - n
	if startIndex < 0 {
		startIndex = 0 // If n is larger than total records, return all records
	}

	return records[startIndex:], nil
}
```

## 2.4. Containerization with Docker

With the Go application code complete, the next step is to containerize it using Docker. This involves creating a Dockerfile for building the image and a docker-compose.yml file for defining the service.

### 2.4.1. Multi-Stage Dockerfile

A multi-stage build is a best practice for compiling Go applications into Docker images. It creates a small, secure final image by separating the build environment (which contains the Go toolchain and source code) from the runtime environment (which contains only the compiled binary).

Create a file named Dockerfile in the project root:

```dockerfile
# Dockerfile

# --- Stage 1: Builder ---
# Use the official Go image with an Alpine Linux base. Alpine is a lightweight
# Linux distribution, which helps keep the build stage smaller.
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies.
# This is done as a separate step to leverage Docker's layer caching.
# Dependencies will only be re-downloaded if these files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application.
# - CGO_ENABLED=0 disables CGO, creating a statically linked binary.
# -o main specifies the output file name.
# -ldflags="-s -w" strips debugging information, reducing the binary size.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

# --- Stage 2: Final Image ---
# Use a minimal Alpine image for the final stage. It's much smaller than
# the Go development image.
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Alpine Linux does not include root CA certificates by default.
# These are necessary for making secure HTTPS connections to other services.
# It's a best practice to include them.
RUN apk --no-cache add ca-certificates

# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/main .

# The application will read the CSV file from a /data directory.
# This directory will be provided by a Docker volume at runtime.
# We don't create it here, but it's good practice to document its expected presence.

# Expose the port the application will run on.
EXPOSE 8080

# The command to run when the container starts.
CMD ["./main"]
```

### 2.4.2. Docker Compose File (docker-compose.yml)

The docker-compose.yml file defines the service, its build context, port mappings, volume mounts, and environment variables. This file makes deployment reproducible and simple.

Create a file named docker-compose.yml in the project root:

```yaml
# docker-compose.yml
version: '3.8'

services:
  claude-connector:
    # Build the image from the Dockerfile in the current directory
    build: .
    # Name the container for easier identification
    container_name: claude-connector-service
    # Map port 8080 on the host to port 8080 in the container
    ports:
      - "8080:8080"
    # Mount the local ./data directory to the /data directory inside the container.
    # This makes the medical_data.csv file available to the application.
    # The ':ro' flag makes the volume read-only for added security.
    volumes:
      - ./data:/data:ro
    # Load environment variables from a .env file in the same directory
    env_file:
      - .env
    # Automatically restart the container unless it is explicitly stopped.
    # This improves the resilience of the service.
    restart: unless-stopped
```

Finally, create a .env file in the root directory to store your secrets. This file should never be committed to version control.

```env
# .env
# Port for the MCP server to listen on
MCP_SERVER_PORT=8080

# A strong, randomly generated secret key for authentication
API_SECRET_KEY=your-super-secret-and-long-random-string

# The path to the CSV file *inside the container*
CSV_FILE_PATH=/data/medical_data.csv
```

## 2.5. Deployment and Configuration

This section details the final steps to deploy the containerized application to your server and expose it securely to the internet via Cloudflare.

### 2.5.1. Cloudflare DNS Configuration

- Log in to your Cloudflare dashboard.
- Select the domain you wish to use.
- Navigate to the DNS > Records section.
- Click Add record.
- Configure the record with the following settings:
  - Type: A
  - Name: The subdomain you want to use (e.g., claude-connector).
  - IPv4 address: The public IP address of your server.
  - Proxy status: Ensure it is set to Proxied (orange cloud icon). This activates Cloudflare's security and TLS services.
- Click Save.

DNS propagation is typically fast but may take a few minutes.

### 2.5.2. Portainer Deployment

- Transfer the entire claude-connector project directory (including the Dockerfile, docker-compose.yml, .env file, and the data directory with your CSV file) to your server.
- Log in to your Portainer instance.
- From the left-hand menu, select Stacks.
- Click Add stack.
- Provide a name for the stack, such as claude-connector-stack.
- In the build method section, select Web editor.
- Copy the contents of your docker-compose.yml file and paste them into the editor.
- Scroll down to the Environment variables section. Click Add environment variable three times and manually enter the key-value pairs from your .env file (MCP_SERVER_PORT, API_SECRET_KEY, CSV_FILE_PATH). Storing secrets this way is more secure than relying on the .env file in some Portainer setups.
- At the bottom of the page, click Deploy the stack.

Portainer will now pull the necessary base images, build your application image, and start the container according to your docker-compose.yml configuration. You can view the container's logs by navigating to Containers, clicking on your claude-connector-service container, and then clicking the Logs icon.

### 2.5.3. Connecting to Claude.ai

- Navigate to the Claude.ai website and log in to your Pro or Max account.
- Go to Settings by clicking your profile icon.
- Select the Connectors (or Custom Connectors) section.
- Scroll down and click Add custom connector.
- A dialog will appear prompting for the server URL. Enter the full, secure URL you configured in Cloudflare, including the /mcp path: https://claude-connector.yourdomain.com/mcp.
- Claude will attempt to connect. It may prompt for authentication. Depending on the interface, you will need to specify that authentication is done via a Bearer token in the Authorization header. Enter the same secret key you defined in the API_SECRET_KEY environment variable.
- Once successfully connected, the connector and its get_last_n_records tool will be available for use in your Claude conversations.

## 2.6. System Troubleshooting Guide

If you encounter issues, follow these steps to diagnose and resolve them.

### Error in Claude: "Connector failed to connect"

- **Check Portainer Logs**: The first step is always to check the application logs. In Portainer, navigate to the container's log view. Look for any error messages at startup, such as "FATAL: API_SECRET_KEY not set" or panics related to file paths or permissions.
- **Verify Network Path with curl**: Use a command-line tool like curl from your local machine (not the server) to test the entire request path, bypassing the Claude interface. This helps isolate whether the issue is with Claude or your service.

```bash
curl -v -X POST \
  -H "Authorization: Bearer your-super-secret-and-long-random-string" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0", "method":"initialize", "id":1}' \
  https://claude-connector.yourdomain.com/mcp
```

A successful response will be a JSON object with a result key. An error will provide more clues. The -v (verbose) flag shows details of the TLS handshake and HTTP headers.

### Cloudflare Displays a 5xx Error (e.g., 502 Bad Gateway, 521 Web server is down)

- **Check Container Status**: In Portainer, verify that the claude-connector-service container is in a running state. If it's stopped or restarting, check the logs for crash information.
- **Check Server Firewall**: Ensure that your server's firewall is not blocking incoming traffic on the port you mapped (e.g., 8080) from Cloudflare's IP ranges. For simplicity during testing, you can temporarily allow all traffic on that port.
- **Verify Port Mapping**: Double-check the ports section in your docker-compose.yml and ensure the host port is correctly configured and not in use by another application on the server.

### Application Responds with 401 Unauthorized

- **Mismatched Keys**: This error means the Bearer token sent in the request does not match the API_SECRET_KEY expected by the server.
- **Verify Server Key**: In Portainer, go to the stack editor or container details and verify the value of the API_SECRET_KEY environment variable.
- **Verify Client Key**: In the Claude connector settings, re-enter the API key, ensuring there are no extra spaces or characters. If testing with curl, double-check the key in the Authorization header.

### Tool get_last_n_records Does Not Appear in Claude

- **MCP Initialization Failure**: This indicates that the initial handshake between Claude and your server is failing, or the server is not correctly advertising its tools.
- **Check initialize Response**: Use the curl command from above to test the initialize method. The response should contain information about the server.
- **Check tools/list Logic**: The mcp-golang library automatically handles the tools/list method based on the tools you register. Check the Portainer logs for any errors that might occur during the server.RegisterTool call in mcp_handler.go. An error here would prevent the tool from being advertised.
- **File Path Error**: The tool might fail if it cannot find the CSV file. The error "could not open csv file" in the logs points to an issue with the CSV_FILE_PATH environment variable or the Docker volume mount. Ensure the path /data/medical_data.csv is correct and that the data directory on the host contains the file.