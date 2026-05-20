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

type GetPatentClaimsInput struct {
	PublicationNumber string `json:"publication_number" jsonschema:"the patent publication number, e.g. US-7650331-B1"`
}

var GetPatentClaimsTool = &mcp.Tool{
	Name:        "get_patent_claims",
	Description: "Fetch the full claims text of a US patent by publication number. Claims are the legal scope of the patent.",
}

func GetPatentClaimsHandler() func(context.Context, *mcp.CallToolRequest, GetPatentClaimsInput) (*mcp.CallToolResult, any, error) {
	client, err := bq.NewClient(context.Background(), os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		logger.Log.Fatal("get_patent_claims: failed to create BigQuery client", zap.Error(err))
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, input GetPatentClaimsInput) (*mcp.CallToolResult, any, error) {
		claims, err := client.GetPatentClaims(ctx, input.PublicationNumber)
		if err != nil {
			return nil, nil, fmt.Errorf("get claims failed: %w", err)
		}

		out, err := json.Marshal(map[string]any{
			"publication_number": input.PublicationNumber,
			"claims":             claims,
		})
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
