package storage

import (
	"os"
	"path/filepath"
)

func ToolsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "GRTBox", "tools"), nil
}

func ExtractedToolsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "GRTBox", "extracted_tools"), nil
}

func BootstrapStatePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "GRTBox", "first_run_bootstrap.json"), nil
}
