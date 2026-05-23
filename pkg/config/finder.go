package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type WorkspaceConfig struct {
	Port          int    `json:"port"`
	WorkspacePath string `json:"workspacePath"`
	Pid           int    `json:"pid"`
}

type RegistryEntry struct {
	ConfigPath    string `json:"configPath"`
	WorkspacePath string `json:"workspacePath"`
	WorkspaceName string `json:"workspaceName"`
	Port          int    `json:"port"`
}

type Registry struct {
	ActiveInstances []RegistryEntry `json:"activeInstances"`
}

func logInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[CLI] "+format+"\n", args...)
}

func isPidAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, sending signal 0 checks if the process exists.
	// We handle errors below, but process finding doesn't typically error on Unix.
	if err := p.Signal(os.Signal(nil)); err == nil {
		return true
	}
	// Fallback to true if signaling fails (e.g. Windows)
	return true
}

func findWorkspaceConfig() (*WorkspaceConfig, string) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, ""
	}

	root := filepath.VolumeName(currentDir) + string(os.PathSeparator)

	for currentDir != root && currentDir != "." {
		configPath := filepath.Join(currentDir, ".mcp-debug-tools", "config.json")
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err == nil {
				var config WorkspaceConfig
				if err := json.Unmarshal(data, &config); err == nil {
					if isPidAlive(config.Pid) {
						logInfo("Workspace config found: %s", currentDir)
						return &config, configPath
					} else {
						logInfo("Stale config ignored: %s", configPath)
					}
				}
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	return nil, ""
}

func findFromGlobalRegistry() []RegistryEntry {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	registryPath := filepath.Join(homeDir, ".mcp-debug-tools", "active-configs.json")
	if _, err := os.Stat(registryPath); err != nil {
		return nil
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		logInfo("Failed to read global registry: %v", err)
		return nil
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil
	}

	var activeInstances []RegistryEntry
	for _, entry := range registry.ActiveInstances {
		if _, err := os.Stat(entry.ConfigPath); err == nil {
			configData, err := os.ReadFile(entry.ConfigPath)
			if err == nil {
				var config WorkspaceConfig
				if err := json.Unmarshal(configData, &config); err == nil {
					if isPidAlive(config.Pid) {
						activeInstances = append(activeInstances, entry)
					}
				}
			}
		}
	}

	return activeInstances
}

func FindVSCodeInstance() (int, string, bool) {
	logInfo("Auto-discovering VSCode instance...")

	wsConfig, _ := findWorkspaceConfig()
	if wsConfig != nil {
		logInfo("✅ Workspace VSCode found - Port: %d", wsConfig.Port)
		return wsConfig.Port, wsConfig.WorkspacePath, true
	}

	logInfo("Not found in Workspace, checking global registry...")
	instances := findFromGlobalRegistry()

	if len(instances) == 0 {
		logInfo("❌ No active VSCode instance found")
		return 0, "", false
	}

	if len(instances) == 1 {
		instance := instances[0]
		logInfo("✅ Single VSCode found - %s (Port: %d)", instance.WorkspaceName, instance.Port)
		return instance.Port, instance.WorkspacePath, true
	}

	logInfo("🔍 Found %d active VSCode instances:", len(instances))
	for i, inst := range instances {
		logInfo("  %d. %s (Port: %d)", i+1, inst.WorkspaceName, inst.Port)
	}

	selected := instances[0]
	logInfo("⚡ First instance auto-selected: %s", selected.WorkspaceName)

	return selected.Port, selected.WorkspacePath, true
}
