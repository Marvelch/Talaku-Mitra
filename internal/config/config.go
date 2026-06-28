package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessExpiry  string
	JWTRefreshExpiry string

	// JWT secret yang sama dengan talaku-microservice, dipakai untuk validasi
	// token customer dan driver yang diterbitkan oleh main service.
	MainJWTSecret string

	ServerPort string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	PasswordResetCodeExpiry time.Duration

	WAGatewayURL     string
	ZenzivaUserKey   string
	ZenzivaPassKey   string
	AppBrandName     string
	BackofficeSecret string
}

var AppConfig *Config

func Load() {
	// Cari .env dari direktori saat ini, lalu naik hingga 3 level
	loaded := false
	for _, path := range []string{".env", "../.env", "../../.env", "../../../.env"} {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded .env from: %s", path)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASS", ""),
		DBName:     getEnv("DB_NAME", "talaku"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTAccessSecret:  mustGetEnv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: mustGetEnv("JWT_REFRESH_SECRET"),
		JWTAccessExpiry:  getEnv("JWT_ACCESS_EXPIRY", "15m"),
		JWTRefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "168h"),

		MainJWTSecret: mustGetEnv("MAIN_JWT_SECRET"),

		ServerPort: getEnv("SERVER_PORT", "8080"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "no-reply@talaku.app"),

		PasswordResetCodeExpiry: getDuration("PASSWORD_RESET_CODE_EXPIRY", 15*time.Minute),

		WAGatewayURL:     getEnv("WA_GATEWAY_URL", ""),
		ZenzivaUserKey:   getEnv("ZENZIVA_USERKEY", ""),
		ZenzivaPassKey:   getEnv("ZENZIVA_PASSKEY", ""),
		AppBrandName:     getEnv("APP_BRAND_NAME", "Talaku Mitra"),
		BackofficeSecret: getEnv("BACKOFFICE_SECRET", "talaku-mitra-bo-secret"),
	}

	if AppConfig.SMTPHost == "" {
		log.Println("SMTP_HOST tidak diatur — kode reset password akan dicatat ke log, bukan dikirim via email.")
	}
	if AppConfig.WAGatewayURL == "" {
		log.Println("WA_GATEWAY_URL tidak diatur — kode OTP akan dicatat ke log, bukan dikirim via WhatsApp.")
	}
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("FATAL: environment variable %s wajib dikonfigurasi", key)
	}
	return v
}

func GetInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
