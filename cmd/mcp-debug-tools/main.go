package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hwanyong/mcp-debug-tools/pkg/action"
	"github.com/hwanyong/mcp-debug-tools/pkg/config"
	"github.com/hwanyong/mcp-debug-tools/pkg/proxy"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
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
	rootCmd := &cobra.Command{
		Use:     "mcp-debug-tools",
		Short:   "CLI and MCP proxy for VSCode debugging via DAP",
		Version: "1.0.2",
		Run: func(cmd *cobra.Command, args []string) {
			serverUrl := getServerUrl()
			startProxy(serverUrl)
		},
	}

	rootCmd.PersistentFlags().IntVar(&portFlag, "port", 0, "DAP Proxy server port (disables auto discovery)")
	rootCmd.PersistentFlags().StringVar(&domainFlag, "domain", "http://localhost", "DAP Proxy server domain")
	rootCmd.PersistentFlags().BoolVar(&noAutoFlag, "no-auto", false, "Disable automatic VSCode discovery")

	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "Start the stdio MCP proxy (Default)",
		Run: func(cmd *cobra.Command, args []string) {
			serverUrl := getServerUrl()
			startProxy(serverUrl)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available MCP tools and resources from the VSCode extension",
		Run: func(cmd *cobra.Command, args []string) {
			serverUrl := getServerUrl()
			action.ListToolsAndResources(serverUrl)
		},
	}

	callCmd := &cobra.Command{
		Use:   "call [toolName] [argsJson]",
		Short: "Call a specific MCP tool directly and print the JSON result",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			serverUrl := getServerUrl()
			toolName := args[0]
			argsJson := ""
			if len(args) > 1 {
				argsJson = args[1]
			}
			action.CallTool(serverUrl, toolName, argsJson)
		},
	}

	readCmd := &cobra.Command{
		Use:   "read [resourceUri]",
		Short: "Read a specific MCP resource directly and print the JSON result",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serverUrl := getServerUrl()
			resourceUri := args[0]
			action.ReadResource(serverUrl, resourceUri)
		},
	}

	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(callCmd)
	rootCmd.AddCommand(readCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
