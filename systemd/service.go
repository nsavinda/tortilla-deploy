package systemd

import (
	"fmt"
	"os"
	"path/filepath"

	"AutoPuller/config"
)

func GenerateServiceFiles(serviceName string) ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	services, ok := cfg.Services[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found in config", serviceName)
	}

	var createdUnits []string

	for _, service := range services {
		colors := []string{"blue", "green"}
		ports := service.TargetPorts

		if len(colors) != len(ports) {
			return nil, fmt.Errorf("mismatch between number of colors and target ports for service %s", service.Name)
		}

		for i, color := range colors {
			port := ports[i]
			unitName := fmt.Sprintf("%s.%s.service", service.Name, color)
			filePath := filepath.Join("/etc/systemd/system", unitName)

			serviceContent := fmt.Sprintf(`[Unit]
Description=My App Service - %s
After=network.target

[Service]
# ExecStartPre=%s
WorkingDirectory=%s%s
Environment=PORT=%d
ExecStart=%s%s/%s
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
`, color,
				service.PreStartHook,
				service.DeploymentsDir, color,
				port,
				service.DeploymentsDir, color,
				service.Executable,
			)

			if err := os.WriteFile(filePath, []byte(serviceContent), 0644); err != nil {
				return nil, fmt.Errorf("failed to write %s service file: %w", color, err)
			}

			createdUnits = append(createdUnits, unitName)
		}
	}

	return createdUnits, nil
}

func GenerateAllServiceFiles() ([]string, error) {
	var allCreatedUnits []string
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	for serviceName := range cfg.Services {
		createdUnits, err := GenerateServiceFiles(serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate service files for %s: %w", serviceName, err)
		}
		allCreatedUnits = append(allCreatedUnits, createdUnits...)
	}

	return allCreatedUnits, nil
}
