package fcm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const fcmEndpoint = "https://fcm.googleapis.com/fcm/send"

type Service struct {
	serverKey string
}

func New() *Service {
	return &Service{serverKey: os.Getenv("FCM_SERVER_KEY")}
}

func (s *Service) Send(token, title, body string, data map[string]string) error {
	if s.serverKey == "" || token == "" {
		return nil
	}
	return s.send(token, title, body, data)
}

// SendToTopic mengirim notifikasi FCM ke semua perangkat yang subscribe topic.
// topic harus tanpa prefix "/topics/" — method ini menambahkannya sendiri.
func (s *Service) SendToTopic(topic, title, body string, data map[string]string) error {
	if s.serverKey == "" || topic == "" {
		return nil
	}
	return s.send("/topics/"+topic, title, body, data)
}

func (s *Service) send(to, title, body string, data map[string]string) error {
	payload := map[string]interface{}{
		"to": to,
		"notification": map[string]string{
			"title": title,
			"body":  body,
			"sound": "default",
		},
		"data":     data,
		"priority": "high",
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fcmEndpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "key="+s.serverKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("FCM responded with status %d", resp.StatusCode)
	}
	return nil
}
