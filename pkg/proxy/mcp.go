package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func logInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[CLI] "+format+"\n", args...)
}

func CreateMcpProxy(serverUrl string) (*server.MCPServer, *client.Client, error) {
	logInfo("🔗 Creating HTTP client: %s", serverUrl)

	mcpClient, err := client.NewSSEMCPClient(serverUrl)
	if err != nil {
		return nil, nil, err
	}

	logInfo("📡 Connected HTTP Transport")

	proxyServer := server.NewMCPServer("dap-proxy-client", "1.0.0")

	logInfo("🔧 Fetching tool list...")
	ctx := context.Background()

	toolsResp, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools: %w", err)
	}

	logInfo("📋 Found tools: %d", len(toolsResp.Tools))

	for _, tool := range toolsResp.Tools {
		logInfo("  - Registering tool: %s", tool.Name)

		toolCopy := tool

		proxyServer.AddTool(toolCopy, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

			argsBytes, _ := json.Marshal(request.Params.Arguments)
			logInfo("🛠️ Calling tool: %s - %s", toolCopy.Name, string(argsBytes))

			startTime := time.Now()

			ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			res, err := mcpClient.CallTool(ctxTimeout, request)

			duration := time.Since(startTime)
			if err != nil {
				logInfo("❌ Tool call failed: %s - %s (%d ms)", toolCopy.Name, err.Error(), duration.Milliseconds())
				return nil, err
			}

			logInfo("✅ Tool call completed: %s (%d ms)", toolCopy.Name, duration.Milliseconds())
			return res, nil
		})
	}

	logInfo("📚 Fetching resource list...")
	resourcesResp, err := mcpClient.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list resources: %w", err)
	}

	logInfo("📋 Found resources: %d", len(resourcesResp.Resources))

	for _, resource := range resourcesResp.Resources {
		logInfo("  - Registering resource: %s: %s", resource.Name, resource.Description)
		resourceCopy := resource

		proxyServer.AddResource(resourceCopy, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			logInfo("📖 Reading resource: %s", resourceCopy.Name)
			res, err := mcpClient.ReadResource(ctx, mcp.ReadResourceRequest{
				Params: mcp.ReadResourceParams{
					URI: request.Params.URI,
				},
			})
			if err != nil {
				return nil, err
			}
			logInfo("✅ Resource read completed: %s", resourceCopy.Name)
			return res.Contents, nil
		})
	}

	logInfo("🎯 MCP Proxy server is ready")

	return proxyServer, mcpClient, nil
}
