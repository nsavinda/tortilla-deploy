package deploy

import (
	"encoding/json"
	"fmt"
	"io"
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

type PushEvent struct {
	Ref        string `json:"ref"`
	HeadCommit struct {
		ID string `json:"id"`
	} `json:"head_commit"`
	Repository struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}

// WebhookHandler handles the webhook POST request
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	var event PushEvent
	// Unmarshal the body into PushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	// Load the configuration from the YAML file
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "config loading error", http.StatusInternalServerError)
		return
	}

	// Validate the repository and branch
	if event.Repository.CloneURL != cfg.Repo.CloneURL || event.Ref != "refs/heads/"+cfg.Repo.Branch {
		http.Error(w, "unauthorized repository or branch", http.StatusForbidden)
		return
	}

	sha := event.HeadCommit.ID
	if sha == "" {
		http.Error(w, "commit SHA missing", http.StatusBadRequest)
		return
	}

	// Get the currently active service
	current := util.GetActiveService()

	// Get the next service name based on current active service
	next := util.GetNextService(current)

	// Define the directory for the deployment
	// dir := fmt.Sprintf("deployments/%s", sha)
	dir := fmt.Sprintf("%s%s", cfg.Service.ClonePath, next)

	// Clone or pull the repository
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Clone if it doesn't exist
		cmd := exec.Command("git", "clone", cfg.Repo.CloneURL, dir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			http.Error(w, "repository clone failed", http.StatusInternalServerError)
			return
		}
	} else {
		// If the directory exists, just pull the latest changes
		cmd := exec.Command("git", "-C", dir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			http.Error(w, "repository pull failed", http.StatusInternalServerError)
			return
		}
	}

	// Run the pre-start hook if defined
	if cfg.Service.PreStartHook != "" {
		// %i in the PreStartHook is replaced with the next service name
		// next := util.GetNextService(util.GetActiveService())
		preStartHook := strings.ReplaceAll(cfg.Service.PreStartHook, "%i", next)
		cmd := exec.Command(preStartHook)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("pre-start hook failed:", err)
			http.Error(w, "pre-start hook failed", http.StatusInternalServerError)
			return
		}
	}

	// Format unit name based on service name and commit SHA
	unit := fmt.Sprintf("%s@%s.service", cfg.Service.Name, next)

	// Restart the new service
	if err := systemd.RestartService(unit); err != nil {
		log.Println("Failed to restart the next service:", err)
		http.Error(w, "failed to restart service", http.StatusInternalServerError)
		return
	}

	// Switch traffic to the new service
	// forwarder := traffic.NewForwarder()
	// if err := forwarder.Start(cfg.Service.SrcPort, cfg.Service.DestPorts[int(next[0]-'0')-1]); err != nil {
	// 	log.Println("Failed to start traffic forwarder:", err)
	// }
	// log.Println("Traffic switched to:", cfg.Service.DestPorts[int(next[0]-'0')-1])
	if err := traffic.UpdateIPTables(cfg.Service.ListenPort, cfg.Service.DestPorts[int(next[0]-'0')-1]); err != nil {
		log.Println("Failed to update iptables:", err)
		http.Error(w, "failed to update iptables", http.StatusInternalServerError)
		return
	}

	// Stop the current active service
	currentUnit := fmt.Sprintf("%s@%s.service", cfg.Service.Name, current)
	if err := systemd.StopService(currentUnit); err != nil {
		log.Println("Failed to stop the current service:", err)
		http.Error(w, "failed to stop current service", http.StatusInternalServerError)

		return
	}

	// Update the active service to the new one
	if err := util.SetActiveService(next); err != nil {
		log.Println("Failed to set active service:", err)
		http.Error(w, "failed to set active service", http.StatusInternalServerError)
		return
	}

	log.Println("Deployment successful. Switched to:", next)
	w.WriteHeader(http.StatusOK)
}
