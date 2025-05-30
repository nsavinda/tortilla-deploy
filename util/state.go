package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const stateFileName = "active_service.txt"

// getStateFile returns the full path to the state file in /var/lib/<service-name>/
func getStateFile(serviceName string) (string, error) {
	// cfg, err := config.Load()
	// if err != nil {
	// 	return "", err
	// }
	stateDir := filepath.Join("/var/lib", serviceName)
	return filepath.Join(stateDir, stateFileName), nil
}

// GetActiveService returns the current active service from the state file.
func GetActiveService(serviceName string) string {
	stateFile, err := getStateFile(serviceName)
	fmt.Println("State file path:", stateFile)
	if err != nil {
		return "blue"
	}
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return "blue"
	}
	return strings.TrimSpace(string(data))
}

// GetNextService returns the alternate service name.
func GetNextService(current string) string {
	if current == "blue" {
		return "green"
	}
	return "blue"
}

// SetActiveService sets the active service in the state file.
func SetActiveService(serviceName string, next string) error {
	fmt.Println("Setting active service to:", serviceName)
	stateFile, err := getStateFile(serviceName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(stateFile, []byte(next), 0644)
}
