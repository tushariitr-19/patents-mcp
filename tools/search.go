package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	bq "github.com/tushariitr-19/patents-mcp/bigquery"
	"github.com/tushariitr-19/patents-mcp/logger"
)

type SearchPatentsInput struct {
	Query string `json:"query" jsonschema:"the keyword, technology area, or inventor name to search for"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results to return, default 10, max 50"`
}

var SearchPatentsTool = &mcp.Tool{
	Name:        "search_patents",
	Description: "Search patents by keyword, inventor, or technology area using Google Patents public dataset",
}

func SearchPatentsHandler() func(context.Context, *mcp.CallToolRequest, SearchPatentsInput) (*mcp.CallToolResult, any, error) {
	client, err := bq.NewClient(context.Background(), os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		logger.Log.Error("search_patents: failed to create BigQuery client", zap.Error(err))
		return func(ctx context.Context, req *mcp.CallToolRequest, input SearchPatentsInput) (*mcp.CallToolResult, any, error) {
			return nil, nil, fmt.Errorf("BigQuery client not initialized: %w", err)
		}
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchPatentsInput) (*mcp.CallToolResult, any, error) {
		if input.Limit == 0 {
			input.Limit = 10
		}
		if input.Limit > 50 {
			input.Limit = 50
		}

		patents, err := client.SearchPatents(ctx, input.Query, input.Limit)
		if err != nil {
			return nil, nil, fmt.Errorf("search failed: %w", err)
		}

		out, err := json.Marshal(patents)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(out)},
			},
		}, nil, nil
	}
}
