package util

import (
	"os"
	"strings"
)

const stateFile = "active_service.txt"

func GetActiveService() string {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return "blue" // default
	}
	return strings.TrimSpace(string(data))
}

func GetNextService(current string) string {
	if current == "blue" {
		return "green"
	}
	return "blue"
}

func SetActiveService(svc string) error {
	return os.WriteFile(stateFile, []byte(svc), 0644)
}
