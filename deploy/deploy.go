package deploy

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

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
	// if cloneURL != cfg.Repository.URL || ref != "refs/heads/"+cfg.Repository.Branch {
	// 	return http.StatusForbidden, fmt.Errorf("unauthorized repository or branch")
	// }
	// Check if the clone URL matches any of the configured services
	var serviceFound bool
	var serviceName string
	for _, services := range cfg.Services {
		for _, service := range services {
			if service.Repository.URL == cloneURL && ref == "refs/heads/"+service.Repository.Branch {
				serviceFound = true
				serviceName = service.Name
				log.Printf("Deploying service: %s from branch: %s", serviceName, service.Repository.Branch)
				break
			}
		}
		if serviceFound {
			break
		}
	}
	if !serviceFound {
		return http.StatusForbidden, fmt.Errorf("unauthorized repository or branch")
	}

	if sha == "" {
		return http.StatusBadRequest, fmt.Errorf("commit SHA missing")
	}

	current := util.GetActiveService(serviceName)
	next := util.GetNextService(current)
	dir := fmt.Sprintf("%s%s", cfg.Services[serviceName][0].DeploymentsDir, next)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create deployment directory: %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := run("git", "clone", cfg.Services[serviceName][0].Repository.URL, dir); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("repository clone failed: %v", err)
		}
	} else {
		if err := run("git", "-C", dir, "pull"); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("repository pull failed: %v", err)
		}
	}

	if cfg.Services[serviceName][0].PreStartHook != "" {
		// hook := strings.ReplaceAll(cfg.Service.PreStartHook, "%i", next)
		// hook := cfg.Service.DeploymentsDir + cfg.Service.PreStartHook
		// use path combine
		hook := path.Join(cfg.Services[serviceName][0].DeploymentsDir, next, cfg.Services[serviceName][0].PreStartHook)

		cmd := exec.Command(hook)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("pre-start hook failed:", err)
			return http.StatusInternalServerError, fmt.Errorf("pre-start hook failed: %v", err)
		}
	}

	unit := fmt.Sprintf("%s.%s.service", cfg.Services[serviceName][0].Name, next)
	if err := systemd.RestartService(unit); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to restart new service: %v", err)
	}

	var destPort, prevPort int
	if next == "blue" {
		destPort = cfg.Services[serviceName][0].TargetPorts[0]
		prevPort = cfg.Services[serviceName][0].TargetPorts[1]
	} else {
		destPort = cfg.Services[serviceName][0].TargetPorts[1]
		prevPort = cfg.Services[serviceName][0].TargetPorts[0]
	}

	if err := traffic.UpdateIPTables(cfg.Services[serviceName][0].ListenPort, destPort, prevPort); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("iptables update failed: %v", err)
	}

	currentUnit := fmt.Sprintf("%s.%s.service", cfg.Services[serviceName][0].Name, current)
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
