package workspace

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type ClientSettingsType struct {
	VirtualEnvPath string
	WorkspaceRoot  string
}

var ClientSettings ClientSettingsType

func (c *ClientSettingsType) ModulesPath() []string {
	paths := []string{}
	// Add the default workspace folder
	paths = append(paths, c.WorkspaceRoot)
	// Add the virtual environment path if it is set
	if c.VirtualEnvPath != "" {
		virtualEnvLibPath := filepath.Join(c.VirtualEnvPath, "lib")
		entries, err := os.ReadDir(virtualEnvLibPath)
		if err != nil {
			slog.Error("Error reading directory", slog.String("path", virtualEnvLibPath), slog.Any("error", err))
		} else {
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "python") {
					paths = append(paths, filepath.Join(virtualEnvLibPath, entry.Name(), "site-packages"))
				}
			}
		}
	}
	return paths
}
