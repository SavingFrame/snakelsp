// Package workspace handles Python workspace analysis and symbol extraction.
// It provides functionality for parsing Python files, extracting symbols,
// managing workspace state, and handling virtual environment paths for the LSP server.
package workspace

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ClientSettingsType struct {
	VirtualEnvPath string
	WorkspaceRoot  string
	ModulesPath    []string
}

var ClientSettings ClientSettingsType

func SetClientSettings(virtualEnvPath, workspaceRoot string) *ClientSettingsType {
	settings := &ClientSettingsType{
		VirtualEnvPath: virtualEnvPath,
		WorkspaceRoot:  workspaceRoot,
		ModulesPath:    calculateModulesPath(virtualEnvPath, workspaceRoot),
	}
	ClientSettings = *settings
	return settings
}

func calculateModulesPath(virtualEnvPath, worspaceRoot string) []string {
	paths := []string{}
	virtualEnvModulesPaths, err := getVirtualEnvModulesPath(virtualEnvPath)
	if err != nil {
		slog.Error("Error getting virtual environment modules path", slog.Any("error", err))
	} else {
		paths = append(paths, virtualEnvModulesPaths...)
	}
	pythonModulesPaths := getPythonModulesPath(virtualEnvPath)
	paths = append(paths, pythonModulesPaths...)
	paths = append(paths, filepath.Join(worspaceRoot))
	slog.Info("Calculated modules paths", slog.Any("paths", paths))
	return paths
}

func getVirtualEnvModulesPath(virtualEnvPath string) ([]string, error) {
	paths := []string{}
	if virtualEnvPath == "" {
		return nil, fmt.Errorf("virtual environment path is empty")
	}

	virtualEnvLibPath := filepath.Join(virtualEnvPath, "lib")

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
	return paths, nil
}

func getPythonModulesPath(virtualEnvPath string) []string {
	paths := []string{}
	var originalPythonPath string
	var err error
	if virtualEnvPath == "" {
		slog.Warn("Virtual environment is not set up yet, using default Python paths")
		originalPythonPath, err = exec.LookPath("python")
		if err != nil {
			slog.Error("Error finding Python executable", slog.Any("error", err))
			return paths
		}
	} else {
		binPath := filepath.Join(virtualEnvPath, "bin")
		pythonPath := filepath.Join(binPath, "python")
		originalPythonPath, err = os.Readlink(pythonPath)
		if err != nil {
			slog.Error("Error reading symlink", slog.String("path", pythonPath), slog.Any("error", err))
			return paths
		}
	}
	pythonDir := filepath.Dir(filepath.Dir(originalPythonPath))

	libPath := filepath.Join(pythonDir, "lib")
	entries, err := os.ReadDir(libPath)
	if err != nil {
		slog.Error("Error reading directory", slog.String("path", libPath), slog.Any("error", err))
	} else {
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "python") {
				paths = append(paths, filepath.Join(libPath, entry.Name()))
			}
		}
	}
	return paths
}
