package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"
	"github.com/AzertoxHDW/sentinel/dashboard/backend/storage"
)

type Alerter struct {
	WebhookURL string
}

func NewAlerter(url string) *Alerter {
	return &Alerter{WebhookURL: url}
}

func (a *Alerter) send(message string, config storage.AlertConfig) {
	if !config.Enabled { return }

	if config.Type == "ntfy" {
		req, _ := http.NewRequest("POST", config.NtfyTopic, strings.NewReader(message))
		if config.NtfyUser != "" {
			req.SetBasicAuth(config.NtfyUser, config.NtfyPass)
		}
		http.DefaultClient.Do(req)
	} else if config.Type == "discord" {
		payload := map[string]interface{}{"content": message}
		body, _ := json.Marshal(payload)
		http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(body))
	}
}

func (a *Alerter) SendOfflineAlert(hostname string, agentID string, config storage.AlertConfig) {
	// If alerts are disabled in UI, stop here
	if !config.Enabled {
		return
	}

	message := fmt.Sprintf("❌ ALERT:\nHost %s (%s) has lost connection\nTime: %s", 
		hostname, agentID, time.Now().Local().Format("02 Jan 15:04 MST"))

	// Call the send helper which handles Discord vs ntfy logic
	a.send(message, config)
}

func (a *Alerter) SendOnlineAlert(hostname string, agentID string, config storage.AlertConfig) {
	// If alerts are disabled in UI, stop here
	if !config.Enabled {
		return
	}

	message := fmt.Sprintf("✅ RECOVERY:\nHost %s (%s) has been marked online\nTime: %s", 
		hostname, agentID, time.Now().Local().Format("02 Jan 15:04 MST"))

	// Call the send helper which handles Discord vs ntfy logic
	a.send(message, config)
}