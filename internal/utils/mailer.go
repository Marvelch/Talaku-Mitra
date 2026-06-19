package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"talaku_mitra/internal/config"
)

// SendPasswordResetEmail sends the OTP code to the user's email. If SMTP is not
// configured (local/dev environments), it logs the code instead of sending it.
func SendPasswordResetEmail(toEmail, fullName, code string) error {
	cfg := config.AppConfig

	subject := "Kode Reset Password Talaku Mitra"
	body := fmt.Sprintf(
		"Halo %s,\r\n\r\nGunakan kode berikut untuk mereset password akun Talaku Mitra Anda:\r\n\r\n%s\r\n\r\nKode ini berlaku selama %s. Jika Anda tidak meminta reset password, abaikan email ini.\r\n",
		fullName, code, cfg.PasswordResetCodeExpiry,
	)

	if cfg.SMTPHost == "" {
		log.Printf("[DEV] SMTP belum dikonfigurasi. Kode reset password untuk %s: %s", toEmail, code)
		return nil
	}

	addr := cfg.SMTPHost + ":" + cfg.SMTPPort
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)

	msg := []byte(
		"From: " + cfg.SMTPFrom + "\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" + body,
	)

	return smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{toEmail}, msg)
}
