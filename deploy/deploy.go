package deploy

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"AutoPuller/config"
	"AutoPuller/systemd"
	"AutoPuller/traffic"
	"AutoPuller/util"
)

func Execute(cloneURL, ref, sha string) (int, error) {
	cfg, err := config.Load()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("config loading error: %v", err)
	}

	// Validate origin
	if cloneURL != cfg.Repository.URL || ref != "refs/heads/"+cfg.Repository.Branch {
		return http.StatusForbidden, fmt.Errorf("unauthorized repository or branch")
	}

	if sha == "" {
		return http.StatusBadRequest, fmt.Errorf("commit SHA missing")
	}

	current := util.GetActiveService()
	next := util.GetNextService(current)
	dir := fmt.Sprintf("%s%s", cfg.Service.DeploymentsDir, next)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := run("git", "clone", cfg.Repository.URL, dir); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("repository clone failed: %v", err)
		}
	} else {
		if err := run("git", "-C", dir, "pull"); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("repository pull failed: %v", err)
		}
	}

	if cfg.Service.PreStartHook != "" {
		hook := strings.ReplaceAll(cfg.Service.PreStartHook, "%i", next)
		cmd := exec.Command(hook)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("pre-start hook failed:", err)
			return http.StatusInternalServerError, fmt.Errorf("pre-start hook failed: %v", err)
		}
	}

	unit := fmt.Sprintf("%s.%s.service", cfg.Service.Name, next)
	if err := systemd.RestartService(unit); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to restart new service: %v", err)
	}

	var destPort, prevPort int
	if next == "blue" {
		destPort = cfg.Service.TargetPorts[0]
		prevPort = cfg.Service.TargetPorts[1]
	} else {
		destPort = cfg.Service.TargetPorts[1]
		prevPort = cfg.Service.TargetPorts[0]
	}

	if err := traffic.UpdateIPTables(cfg.Service.ListenPort, destPort, prevPort); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("iptables update failed: %v", err)
	}

	currentUnit := fmt.Sprintf("%s.%s.service", cfg.Service.Name, current)
	if err := systemd.StopService(currentUnit); err != nil {
		log.Println("Failed to stop current service:", err)
	}

	if err := util.SetActiveService(next); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update active service: %v", err)
	}

	log.Println("Deployment successful. Switched to:", next)
	return http.StatusOK, nil
}

func run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
