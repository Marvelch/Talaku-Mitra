package middleware

import (
	"strings"
	"talaku_mitra/internal/config"
	"talaku_mitra/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// MainServiceClaims adalah struktur claims JWT dari talaku-microservice.
// Sama persis dengan common.Claims di main service.
type MainServiceClaims struct {
	UID        string `json:"uid"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	ClientType string `json:"client_type"` // "user" | "driver" | "admin"
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func parseMainToken(tokenStr string) (*MainServiceClaims, error) {
	claims := &MainServiceClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.AppConfig.MainJWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

// CustomerAuthRequired memvalidasi Bearer token customer dari talaku-microservice.
// Set context: "customerUID", "customerEmail", "customerFullName"
func CustomerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Token tidak ditemukan.", nil)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, 401, "Format token tidak valid.", nil)
			c.Abort()
			return
		}
		claims, err := parseMainToken(parts[1])
		if err != nil || claims.ClientType != "user" {
			response.Error(c, 401, "Token customer tidak valid atau sudah kadaluarsa.", nil)
			c.Abort()
			return
		}
		c.Set("customerUID", claims.UID)
		c.Set("customerEmail", claims.Email)
		c.Set("customerFullName", claims.FullName)
		c.Next()
	}
}

// DriverAuthRequired memvalidasi Bearer token driver dari talaku-microservice.
// Set context: "driverUID", "driverEmail"
func DriverAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Token tidak ditemukan.", nil)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, 401, "Format token tidak valid.", nil)
			c.Abort()
			return
		}
		claims, err := parseMainToken(parts[1])
		if err != nil || claims.ClientType != "driver" {
			response.Error(c, 401, "Token driver tidak valid atau sudah kadaluarsa.", nil)
			c.Abort()
			return
		}
		c.Set("driverUID", claims.UID)
		c.Set("driverEmail", claims.Email)
		c.Next()
	}
}
