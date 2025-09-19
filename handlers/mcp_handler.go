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
