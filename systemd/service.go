package systemd

import (
	"fmt"
	"os"
	"path/filepath"

	"AutoPuller/config"
)

func GenerateServiceFile() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	unitName := fmt.Sprintf("%s@.service", cfg.Service.Name)
	filePath := filepath.Join("/etc/systemd/system", unitName)

	serviceContent := fmt.Sprintf(`[Unit]
Description=My App Service - %%i
After=network.target

[Service]
# ExecStartPre=%s  
WorkingDirectory=%s%%i
Environment=PORT=870%%i
ExecStart=%s%%i/%s
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
`,
		cfg.Service.PreStartHook,
		cfg.Service.ClonePath,
		cfg.Service.ClonePath,
		cfg.Service.ExecFile,
	)

	err = os.WriteFile(filePath, []byte(serviceContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write service file: %w", err)
	}

	return unitName, nil
}
