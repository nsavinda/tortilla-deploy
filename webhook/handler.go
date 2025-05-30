package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"AutoPuller/deploy"
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

func Handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate signature
	secret := os.Getenv("WEBHOOK_SECRET")
	signature := r.Header.Get("X-Hub-Signature-256")
	if !isValidSignature(body, signature, secret) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event PushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	status, err := deploy.Execute(event.Repository.CloneURL, event.Ref, event.HeadCommit.ID)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deployment successful"))
}

func isValidSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
