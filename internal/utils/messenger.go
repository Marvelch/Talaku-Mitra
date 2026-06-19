package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"talaku_mitra/internal/config"
	"time"
)

// SendingOTP sends a WhatsApp OTP message.
// Priority: (1) local WA gateway, (2) Zenziva, (3) log only (dev).
func SendingOTP(phone, otpCode string) error {
	cfg := config.AppConfig
	message := fmt.Sprintf("Kode OTP Anda adalah: %s. Berlaku 5 menit. Jangan berikan kepada siapapun.", otpCode)

	// 1. Coba local WA gateway jika dikonfigurasi
	if cfg.WAGatewayURL != "" {
		if err := sendOTPLocal(phone, message); err == nil {
			return nil
		}
		log.Printf("[WA Gateway] Local gateway gagal, mencoba Zenziva untuk %s", phone)
	}

	// 2. Coba Zenziva jika dikonfigurasi
	if cfg.ZenzivaUserKey != "" && cfg.ZenzivaPassKey != "" {
		return sendOTPZenziva(phone, otpCode)
	}

	// 3. Dev fallback: hanya log
	log.Printf("[DEV] WA gateway tidak dikonfigurasi. Kode OTP untuk %s: %s", phone, otpCode)
	return nil
}

func sendOTPLocal(phone, message string) error {
	cfg := config.AppConfig
	localURL := cfg.WAGatewayURL + "/send-message"

	body, err := json.Marshal(map[string]string{"phone": phone, "message": message})
	if err != nil {
		return err
	}

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(localURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WA gateway lokal mengembalikan status %d", resp.StatusCode)
	}
	return nil
}

func sendOTPZenziva(phone, otpCode string) error {
	cfg := config.AppConfig
	if cfg.ZenzivaUserKey == "" || cfg.ZenzivaPassKey == "" {
		return fmt.Errorf("Zenziva belum dikonfigurasi")
	}

	data := url.Values{}
	data.Set("userkey", cfg.ZenzivaUserKey)
	data.Set("passkey", cfg.ZenzivaPassKey)
	data.Set("to", phone)
	data.Set("brand", cfg.AppBrandName)
	data.Set("otp", otpCode)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://console.zenziva.net/waofficial/api/sendWAOfficial/", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Zenziva mengembalikan status %d", resp.StatusCode)
	}
	return nil
}
