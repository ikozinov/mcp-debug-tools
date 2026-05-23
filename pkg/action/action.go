package action

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func logStderr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[CLI Action] "+format+"\n", args...)
}

func ListToolsAndResources(serverUrl string) {
	logStderr("🔗 Attempting connection: %s", serverUrl)
	mcpClient, err := client.NewSSEMCPClient(serverUrl)
	if err != nil {
		logStderr("❌ Connection failed: %v", err)
		os.Exit(1)
	}
	logStderr("✅ Connection successful")

	logStderr("Fetching tools and resources list...")

	ctx := context.Background()
	toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		logStderr("❌ Error: %v", err)
		os.Exit(1)
	}

	resourcesResult, err := mcpClient.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		logStderr("❌ Error: %v", err)
		os.Exit(1)
	}

	output := map[string]interface{}{
		"tools":     toolsResult.Tools,
		"resources": resourcesResult.Resources,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logStderr("❌ Error marshaling output: %v", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
}

func CallTool(serverUrl string, toolName string, argsStr string) {
	var args map[string]interface{}
	if argsStr != "" {
		if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
			logStderr("❌ Argument JSON parsing error: %s", argsStr)
			os.Exit(1)
		}
	} else {
		args = make(map[string]interface{})
	}

	mcpClient, err := client.NewSSEMCPClient(serverUrl)
	if err != nil {
		logStderr("❌ Error: %v", err)
		fmt.Printf("{\n  \"error\": \"%v\"\n}\n", err)
		os.Exit(1)
	}

	logStderr("🛠️ Calling tool: %s", toolName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	})

	if err != nil {
		logStderr("❌ Error: %v", err)
		fmt.Printf("{\n  \"error\": \"%v\"\n}\n", err)
		os.Exit(1)
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonBytes))
}

func ReadResource(serverUrl string, resourceUri string) {
	mcpClient, err := client.NewSSEMCPClient(serverUrl)
	if err != nil {
		logStderr("❌ Error: %v", err)
		fmt.Printf("{\n  \"error\": \"%v\"\n}\n", err)
		os.Exit(1)
	}

	logStderr("📖 Reading resource: %s", resourceUri)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := mcpClient.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: resourceUri,
		},
	})

	if err != nil {
		logStderr("❌ Error: %v", err)
		fmt.Printf("{\n  \"error\": \"%v\"\n}\n", err)
		os.Exit(1)
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonBytes))
}
