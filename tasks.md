# Sequential Implementation Tasks

This document provides an atomic, sequential list of tasks for a coding agent to execute, creating the entire project from scratch.

## Phase 1: Project Initialization (Tasks 1-4)

### Task 1: Create Project Directory Structure

```bash
mkdir -p claude-connector/data && \
mkdir -p claude-connector/handlers && \
mkdir -p claude-connector/middleware && \
mkdir -p claude-connector/tools
```

### Task 2: Navigate into Project Directory

```bash
cd claude-connector
```

### Task 3: Initialize Go Module

```bash
go mod init github.com/user/claude-connector
```

### Task 4: Install Dependencies

```bash
go get github.com/gin-gonic/gin && go get github.com/metoro-io/mcp-golang
```

## Phase 2: Core Service Implementation (Tasks 5-8)

### Task 5: Create main.go

```bash
cat <<'EOF' > main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/user/claude-connector/handlers"
	"github.com/user/claude-connector/middleware"
)

func main() {
	port := os.Getenv("MCP_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("API_SECRET_KEY")
	if apiKey == "" {
		log.Fatal("FATAL: API_SECRET_KEY environment variable not set.")
	}

	csvPath := os.Getenv("CSV_FILE_PATH")
	if csvPath == "" {
		log.Fatal("FATAL: CSV_FILE_PATH environment variable not set.")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	mcpGroup := router.Group("/mcp")
	{
		mcpGroup.Use(middleware.AuthMiddleware(apiKey))
		mcpGroup.POST("", handlers.MCPHandler(csvPath))
	}

	log.Printf("Starting MCP server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
EOF
```

### Task 6: Create middleware/auth.go

```bash
cat <<'EOF' > middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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
EOF
```

### Task 7: Create handlers/mcp_handler.go

```bash
cat <<'EOF' > handlers/mcp_handler.go
package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
	"github.com/user/claude-connector/tools"
)

type GetLastNRecordsArgs struct {
	Count int `json:"count" jsonschema:"required,description=The number of recent records to retrieve."`
}

func MCPHandler(csvPath string) gin.HandlerFunc {
	transport := http.NewGinTransport()
	server := mcp.NewServer(transport)

	err := server.RegisterTool(
		"get_last_n_records",
		"Retrieves the last N records from the local medical information CSV file.",
		func(args GetLastNRecordsArgs) (*mcp.ToolResponse, error) {
			if args.Count <= 0 {
				return mcp.NewToolResponse(mcp.NewTextContent("Error: count must be a positive integer.")), nil
			}

			records, err := tools.GetLastNRecords(csvPath, args.Count)
			if err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Error: failed to get records: %v", err))), nil
			}

			if len(records) == 0 {
				return mcp.NewToolResponse(mcp.NewTextContent("No records found.")), nil
			}

			var b strings.Builder
			for i, record := range records {
				for j, value := range record {
					b.WriteString(value)
					if j < len(record)-1 {
						b.WriteString(",")
					}
				}
				if i < len(records)-1 {
					b.WriteString("\n")
				}
			}

			return mcp.NewToolResponse(mcp.NewTextContent(b.String())), nil
		},
	)

	if err != nil {
		panic(fmt.Sprintf("Failed to register tool: %v", err))
	}

	return transport.Handler()
}
EOF
```

### Task 8: Create tools/csv_reader.go

```bash
cat <<'EOF' > tools/csv_reader.go
package tools

import (
	"encoding/csv"
	"fmt"
	"os"
)

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
		return [][]string{}, nil
	}

	startIndex := totalRecords - n
	if startIndex < 0 {
		startIndex = 0
	}

	return records[startIndex:], nil
}
EOF
```

## Phase 3: Containerization (Tasks 9-11)

### Task 9: Create Dockerfile

```bash
cat <<'EOF' > Dockerfile
# --- Stage 1: Builder ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

# --- Stage 2: Final Image ---
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
EOF
```

### Task 10: Create docker-compose.yml

```bash
cat <<'EOF' > docker-compose.yml
version: '3.8'

services:
  claude-connector:
    build: .
    container_name: claude-connector-service
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data:ro
    env_file:
      - .env
    restart: unless-stopped
EOF
```

### Task 11: Create .dockerignore

```bash
cat <<'EOF' > .dockerignore
.git
.gitignore
.env
*.md
data/
EOF
```

## Phase 4: Configuration and Data (Tasks 12-13)

### Task 12: Create .env file template

```bash
cat <<'EOF' > .env
# Port for the MCP server to listen on
MCP_SERVER_PORT=8080

# A strong, randomly generated secret key for authentication
API_SECRET_KEY=change-this-to-a-very-long-and-random-string

# The path to the CSV file *inside the container*
CSV_FILE_PATH=/data/medical_data.csv
EOF
```

### Task 13: Create sample medical_data.csv

```bash
cat <<'EOF' > data/medical_data.csv
date,metric,value,unit,notes
2024-09-01,heart_rate,65,bpm,Resting
2024-09-01,blood_pressure,120/80,mmHg,Morning reading
2024-09-02,blood_glucose,95,mg/dL,Fasting
2024-09-03,heart_rate,72,bpm,Resting
2024-09-03,blood_pressure,122/81,mmHg,Evening reading
2024-09-04,weight,180,lbs,Morning weight
2024-09-05,heart_rate,68,bpm,Resting
2024-09-06,blood_pressure,118/79,mmHg,Morning reading
2024-09-07,blood_glucose,105,mg/dL,Post-meal
2024-09-08,heart_rate,75,bpm,After light exercise
EOF
```

## Phase 5: Finalization and Documentation (Tasks 14-20)

### Task 14: Tidy Go Modules

```bash
go mod tidy
```

### Task 15: Format Go Code

```bash
gofmt -w .
```


