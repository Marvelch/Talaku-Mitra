package middleware

import (
	"strings"
	"talaku_mitra/internal/config"
	"talaku_mitra/internal/models"
	"talaku_mitra/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// userFinder is the subset of MitraUserRepository needed by UserExistsRequired.
type userFinder interface {
	FindByUID(uid string) (*models.MitraUser, error)
}

// JWTClaims berisi payload JWT yang diterbitkan untuk akun mitra.
// Semua pengguna service ini adalah mitra, sehingga tidak ada flag tambahan.
type JWTClaims struct {
	UID      string `json:"uid"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Token tidak ditemukan. Silakan login terlebih dahulu.", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, 401, "Format token tidak valid. Gunakan: Bearer <token>", nil)
			c.Abort()
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.AppConfig.JWTAccessSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, 401, "Token tidak valid atau sudah kadaluarsa.", nil)
			c.Abort()
			return
		}

		c.Set("userUID", claims.UID)
		c.Set("userFullName", claims.FullName)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

// UserExistsRequired verifies that the authenticated UID still exists in mitra_users.
// Must be applied AFTER AuthRequired so that "userUID" is already set in the context.
// This prevents stale JWTs (e.g. from hard-deleted accounts) from reaching handlers.
func UserExistsRequired(repo userFinder) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("userUID")
		if uid == "" {
			response.Error(c, 401, "Tidak terautentikasi.", nil)
			c.Abort()
			return
		}
		user, err := repo.FindByUID(uid)
		if err != nil {
			response.Error(c, 500, "Terjadi kesalahan server.", nil)
			c.Abort()
			return
		}
		if user == nil {
			response.Error(c, 401, "Sesi tidak valid. Silakan login kembali.", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
