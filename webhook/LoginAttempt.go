package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type LoginAttempt struct {
	UserGUID string `json:"user_id"`
	IP       string `json:"ip"`
	Event    string `json:"event"`
}

func EditIpWebhook(url string, attempt LoginAttempt) error {
	payload, err := json.Marshal(attempt)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
