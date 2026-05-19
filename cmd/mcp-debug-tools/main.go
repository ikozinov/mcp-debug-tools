package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hwanyong/mcp-debug-tools/pkg/action"
	"github.com/hwanyong/mcp-debug-tools/pkg/config"
	"github.com/hwanyong/mcp-debug-tools/pkg/proxy"
	"github.com/mark3labs/mcp-go/server"
)

var (
	portFlag   int
	domainFlag string
	noAutoFlag bool
)

func getServerUrl() string {
	domain := domainFlag
	port := portFlag
	autoConnect := !noAutoFlag

	if port != 0 {
		if port < 1 || port > 65535 {
			fmt.Fprintf(os.Stderr, "❌ Invalid port number\n")
			os.Exit(1)
		}
		autoConnect = false
	}

	if autoConnect && port == 0 {
		discoveredPort, _, ok := config.FindVSCodeInstance()
		if ok {
			port = discoveredPort
		} else {
			fmt.Fprintf(os.Stderr, "[CLI] ⚠️ VSCode instance not found, using default port\n")
			port = 8890
		}
	} else if port == 0 {
		port = 8890
	}

	return fmt.Sprintf("%s:%d/mcp", domain, port)
}

func logInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[CLI] "+format+"\n", args...)
}

func startProxy(serverUrl string) {
	logInfo("🚀 DAP Proxy MCP Client Starting")
	logInfo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logInfo("🎯 Server URL: %s", serverUrl)
	logInfo("🔗 Attempting HTTP connection to VSCode extension...")

	mcpProxyServer, _, err := proxy.CreateMcpProxy(serverUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error connecting to VSCode extension: %v\n", err)
		os.Exit(1)
	}

	logInfo("📡 Starting stdio transport...")

	stdioServer := server.NewStdioServer(mcpProxyServer)

	logInfo("✅ MCP Client ready!")
	logInfo("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error starting Stdio server: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.IntVar(&portFlag, "port", 0, "DAP Proxy server port (disables auto discovery)")
	flag.StringVar(&domainFlag, "domain", "http://localhost", "DAP Proxy server domain")
	flag.BoolVar(&noAutoFlag, "no-auto", false, "Disable automatic VSCode discovery")

	flag.Parse()

	args := flag.Args()

	command := "proxy"
	if len(args) > 0 {
		command = args[0]
	}

	serverUrl := getServerUrl()

	switch command {
	case "list":
		action.ListToolsAndResources(serverUrl)
	case "call":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: mcp-debug-tools call [toolName] [argsJson]\n")
			os.Exit(1)
		}
		toolName := args[1]
		argsJson := ""
		if len(args) > 2 {
			argsJson = args[2]
		}
		action.CallTool(serverUrl, toolName, argsJson)
	case "read":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: mcp-debug-tools read [resourceUri]\n")
			os.Exit(1)
		}
		resourceUri := args[1]
		action.ReadResource(serverUrl, resourceUri)
	case "proxy":
		startProxy(serverUrl)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
