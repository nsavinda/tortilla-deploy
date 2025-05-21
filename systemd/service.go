package systemd

import (
	"fmt"
	"os"
	"path/filepath"

	"AutoPuller/config"
)

func GenerateServiceFiles() ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	colors := []string{"blue", "green"}
	ports := cfg.Service.DestPorts

	if len(colors) != len(ports) {
		return nil, fmt.Errorf("mismatch between number of colors and destination ports")
	}

	var createdUnits []string

	for i, color := range colors {
		port := ports[i]
		serviceName := fmt.Sprintf("%s.%s.service", cfg.Service.Name, color)
		filePath := filepath.Join("/etc/systemd/system", serviceName)

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
`,
			color,
			cfg.Service.PreStartHook,
			cfg.Service.ClonePath, color,
			port,
			cfg.Service.ClonePath, color,
			cfg.Service.ExecFile,
		)

		if err := os.WriteFile(filePath, []byte(serviceContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s service file: %w", color, err)
		}

		createdUnits = append(createdUnits, serviceName)
	}

	return createdUnits, nil
}
