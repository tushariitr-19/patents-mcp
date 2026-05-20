package main

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)

	cmd := exec.Command("./patents-mcp-server")
	cmd.Env = append(os.Environ(),
		"GOOGLE_APPLICATION_CREDENTIALS="+os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		"GCP_PROJECT_ID="+os.Getenv("GCP_PROJECT_ID"),
	)

	transport := &mcp.CommandTransport{Command: cmd}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// List tools
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("ListTools failed: %v", err)
	}
	for _, t := range tools.Tools {
		log.Printf("tool: %s — %s", t.Name, t.Description)
	}

	// Call search_patents
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "search_patents",
		Arguments: map[string]any{"query": "machine learning", "limit": 3},
	})
	if err != nil {
		log.Fatalf("CallTool failed: %v", err)
	}

	for _, c := range res.Content {
		log.Print(c.(*mcp.TextContent).Text)
	}
}
