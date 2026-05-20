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

type GetPatentInput struct {
	PublicationNumber string `json:"publication_number" jsonschema:"the patent publication number, e.g. US-7650331-B1"`
}

var GetPatentTool = &mcp.Tool{
	Name:        "get_patent",
	Description: "Fetch full details of a patent by its publication number",
}

func GetPatentHandler() func(context.Context, *mcp.CallToolRequest, GetPatentInput) (*mcp.CallToolResult, any, error) {
	client, err := bq.NewClient(context.Background(), os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		logger.Log.Fatal("get_patent: failed to create BigQuery client", zap.Error(err))
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, input GetPatentInput) (*mcp.CallToolResult, any, error) {
		patent, err := client.GetPatent(ctx, input.PublicationNumber)
		if err != nil {
			return nil, nil, fmt.Errorf("get patent failed: %w", err)
		}

		out, err := json.Marshal(patent)
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
