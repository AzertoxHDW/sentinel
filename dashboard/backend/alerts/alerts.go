package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Alerter struct {
	WebhookURL string
}

func NewAlerter(url string) *Alerter {
	return &Alerter{WebhookURL: url}
}

func (a *Alerter) SendOfflineAlert(hostname string, agentID string) {
	if a.WebhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"content": fmt.Sprintf("⚠️ **ALERT:**\nHost `%s` (%s) has lost connection\n**Time:** %s", 
			hostname, agentID, time.Now().Format(time.RFC1123)),
	}

	body, _ := json.Marshal(payload)
	http.Post(a.WebhookURL, "application/json", bytes.NewBuffer(body))
}

func (a *Alerter) SendOnlineAlert(hostname string, agentID string) {
	if a.WebhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"content": fmt.Sprintf("✅ **RECOVERY:**\nHost `%s` (%s) is now online\n**Time:** %s", 
			hostname, agentID, time.Now().Format(time.RFC1123)),
	}

	body, _ := json.Marshal(payload)
	http.Post(a.WebhookURL, "application/json", bytes.NewBuffer(body))
}