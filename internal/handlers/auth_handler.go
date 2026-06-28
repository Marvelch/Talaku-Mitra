package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"talaku_mitra/internal/config"
	"talaku_mitra/internal/middleware"
	"talaku_mitra/internal/models"
	"talaku_mitra/internal/repositories"
	"talaku_mitra/internal/utils"
	"talaku_mitra/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repositories.MitraUserRepository
	otpRepo  *repositories.OtpRepository
}

func NewAuthHandler(userRepo *repositories.MitraUserRepository, otpRepo *repositories.OtpRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, otpRepo: otpRepo}
}

const otpExpiry = 5 * time.Minute

func (h *AuthHandler) sendPhoneOtp(phone string) error {
	if err := h.otpRepo.DeleteUnusedByPhone(phone); err != nil {
		return err
	}

	code, err := utils.GenerateOTP(6)
	if err != nil {
		return err
	}

	otp := &models.OtpVerification{
		Type:      "register_phone",
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(otpExpiry),
	}
	if err := h.otpRepo.Create(otp); err != nil {
		return err
	}

	return utils.SendingOTP(phone, code)
}

func (h *AuthHandler) generateAccessToken(user *models.MitraUser) (string, error) {
	claims := middleware.JWTClaims{
		UID:      user.UID,
		FullName: user.FullName,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.AppConfig.JWTAccessSecret))
}

func (h *AuthHandler) generateRefreshToken(uid string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   uid,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.AppConfig.JWTRefreshSecret))
}

// isVerified returns true only when the pointer is non-nil and true.
func isVerified(v *bool) bool { return v != nil && *v }

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

func generateResetToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// Register godoc
// POST /api/v1/auth/register
// Jika email atau nomor HP sudah ada tapi belum terverifikasi, data lama
// digantikan dengan data registrasi baru (upsert).
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	if msg := utils.ValidatePasswordStrength(req.Password, req.FullName, req.Email); msg != "" {
		response.Error(c, http.StatusBadRequest, msg, nil)
		return
	}

	// ── 1. Periksa email ─────────────────────────────────────────────────
	emailUser, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if emailUser != nil && isVerified(emailUser.IsVerifiedPhone) {
		response.Error(c, http.StatusConflict, "Email sudah terdaftar dan terverifikasi. Silakan login.", nil)
		return
	}

	// ── 2. Periksa nomor HP ──────────────────────────────────────────────
	phoneUser, err := h.userRepo.FindByPhone(req.PhoneNumber)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if phoneUser != nil {
		isDifferentUser := emailUser == nil || phoneUser.UID != emailUser.UID
		if isDifferentUser && isVerified(phoneUser.IsVerifiedPhone) {
			response.Error(c, http.StatusConflict, "Nomor HP sudah digunakan oleh akun terverifikasi.", nil)
			return
		}
		// Nomor HP belum terverifikasi dan milik user berbeda → hapus user lama
		if isDifferentUser {
			_ = h.userRepo.HardDeleteByUID(phoneUser.UID)
		}
	}

	// ── 3. Hash password ─────────────────────────────────────────────────
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memproses password.", nil)
		return
	}

	falseVal := false
	var finalUID string

	// ── 4. Upsert ────────────────────────────────────────────────────────
	if emailUser != nil {
		finalUID = emailUser.UID
		updates := map[string]interface{}{
			"full_name":         req.FullName,
			"phone_number":      req.PhoneNumber,
			"password_hash":     string(hashed),
			"is_active":         true,
			"is_verified_phone": false,
			"refresh_token":     nil,
			"deleted_at":        nil,
		}
		if err := h.userRepo.UpdateCoreFields(finalUID, updates); err != nil {
			response.Error(c, http.StatusInternalServerError, "Gagal memperbarui akun.", nil)
			return
		}
	} else {
		finalUID = uuid.New().String()
		user := &models.MitraUser{
			UID:             finalUID,
			FullName:        req.FullName,
			Email:           req.Email,
			PhoneNumber:     &req.PhoneNumber,
			PasswordHash:    string(hashed),
			IsActive:        true,
			IsVerifiedPhone: &falseVal,
		}
		if err := h.userRepo.Create(user); err != nil {
			response.Error(c, http.StatusInternalServerError, "Gagal membuat akun.", nil)
			return
		}
	}

	// ── 5. Kirim OTP ─────────────────────────────────────────────────────
	if err := h.sendPhoneOtp(req.PhoneNumber); err != nil {
		response.Error(c, http.StatusInternalServerError, "Akun berhasil dibuat, tetapi gagal mengirim kode OTP. Silakan coba kirim ulang.", nil)
		return
	}

	statusCode := http.StatusCreated
	if emailUser != nil {
		statusCode = http.StatusOK
	}

	response.Success(c, statusCode, "Registrasi berhasil. Kode OTP telah dikirim ke nomor WhatsApp Anda.", gin.H{
		"uid":               finalUID,
		"full_name":         req.FullName,
		"email":             req.Email,
		"phone_number":      req.PhoneNumber,
		"is_verified_phone": false,
	})
}

// Login godoc
// POST /api/v1/auth/login
// identifier bisa berupa email atau nomor telepon (08xx / +62xx).
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	var user *models.MitraUser
	var err error

	if strings.Contains(identifier, "@") {
		user, err = h.userRepo.FindByEmail(identifier)
	} else {
		phone := identifier
		if strings.HasPrefix(phone, "0") {
			phone = "+62" + phone[1:]
		} else if !strings.HasPrefix(phone, "+") {
			phone = "+62" + phone
		}
		user, err = h.userRepo.FindByPhone(phone)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}

	const badCredMsg = "Email/nomor HP atau password salah."
	if user == nil || !user.IsActive {
		response.Error(c, http.StatusUnauthorized, badCredMsg, nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, badCredMsg, nil)
		return
	}

	if !isVerified(user.IsVerifiedPhone) {
		response.Error(c, http.StatusForbidden, "Nomor telepon belum diverifikasi. Silakan verifikasi melalui kode OTP yang dikirim ke WhatsApp Anda.", gin.H{
			"phone_number":       user.PhoneNumber,
			"needs_verification": true,
		})
		return
	}

	accessToken, err := h.generateAccessToken(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat access token.", nil)
		return
	}

	refreshTokenStr, err := h.generateRefreshToken(user.UID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat refresh token.", nil)
		return
	}

	if err := h.userRepo.SetRefreshToken(user.UID, &refreshTokenStr); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan token.", nil)
		return
	}
	_ = h.userRepo.UpdateLastLogin(user.UID)

	response.Success(c, http.StatusOK, "Login berhasil.", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshTokenStr,
		"user": gin.H{
			"uid":          user.UID,
			"full_name":    user.FullName,
			"email":        user.Email,
			"phone_number": user.PhoneNumber,
		},
	})
}

// Refresh godoc
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Refresh token wajib diisi.", nil)
		return
	}

	user, err := h.userRepo.FindByRefreshToken(req.RefreshToken)
	if err != nil || user == nil {
		response.Error(c, http.StatusForbidden, "Refresh token tidak valid atau telah dicabut.", nil)
		return
	}

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTRefreshSecret), nil
	})
	if err != nil || !token.Valid {
		response.Error(c, http.StatusForbidden, "Refresh token tidak valid atau sudah kadaluarsa.", nil)
		return
	}

	newAccessToken, err := h.generateAccessToken(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat access token baru.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Token berhasil diperbarui.", gin.H{
		"access_token": newAccessToken,
	})
}

// Logout godoc
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userUID := c.GetString("userUID")
	if err := h.userRepo.SetRefreshToken(userUID, nil); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal logout.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Logout berhasil. Sesi telah dicabut.", nil)
}

// ForgotPassword godoc
// POST /api/v1/auth/forgot-password
// method: "whatsapp" (default) atau "email".
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	method := strings.ToLower(strings.TrimSpace(req.Method))
	if method != "email" {
		method = "whatsapp"
	}

	genericMsg := "Jika email terdaftar, kode OTP telah dikirim."

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if user == nil {
		response.Success(c, http.StatusOK, genericMsg, nil)
		return
	}

	phone := ""
	if user.PhoneNumber != nil {
		phone = *user.PhoneNumber
	}

	code, err := utils.GenerateOTP(6)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat kode OTP.", nil)
		return
	}

	// OTP selalu disimpan berdasarkan nomor HP (digunakan saat verifikasi)
	if phone != "" {
		_ = h.otpRepo.DeleteUnusedByPhoneAndType(phone, "forgot_password")
		otp := &models.OtpVerification{
			Type:      "forgot_password",
			Phone:     phone,
			Code:      code,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}
		if err := h.otpRepo.Create(otp); err != nil {
			response.Error(c, http.StatusInternalServerError, "Gagal menyimpan kode OTP.", nil)
			return
		}
	}

	var phoneHint string
	if method == "email" {
		_ = utils.SendPasswordResetEmail(user.Email, user.FullName, code)
	} else {
		if phone != "" {
			if waErr := utils.SendingOTP(phone, code); waErr != nil {
				// Fallback ke email jika WhatsApp gagal
				_ = utils.SendPasswordResetEmail(user.Email, user.FullName, code)
				method = "email"
			} else {
				phoneHint = maskPhone(phone)
			}
		} else {
			// Tidak ada nomor HP, fallback ke email
			_ = utils.SendPasswordResetEmail(user.Email, user.FullName, code)
			method = "email"
		}
	}

	response.Success(c, http.StatusOK, genericMsg, gin.H{
		"phone_hint": phoneHint,
		"method":     method,
	})
}

// VerifyForgotPasswordOtp godoc
// POST /api/v1/auth/verify-forgot-otp
func (h *AuthHandler) VerifyForgotPasswordOtp(c *gin.Context) {
	var req models.VerifyForgotOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if user == nil || user.PhoneNumber == nil {
		response.Error(c, http.StatusBadRequest, "Kode OTP tidak valid atau sudah kadaluarsa.", nil)
		return
	}

	valid, err := h.otpRepo.ConsumeValidByType(*user.PhoneNumber, req.Code, "forgot_password")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !valid {
		response.Error(c, http.StatusBadRequest, "Kode OTP tidak valid atau sudah kadaluarsa.", nil)
		return
	}

	resetToken, err := generateResetToken()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat token reset.", nil)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(resetToken), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memproses token reset.", nil)
		return
	}
	hashedStr := string(hashed)
	expiresAt := time.Now().Add(15 * time.Minute)

	if err := h.userRepo.SetPasswordResetCode(user.UID, &hashedStr, &expiresAt); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan token reset.", nil)
		return
	}

	response.Success(c, http.StatusOK, "OTP valid. Silakan buat password baru.", gin.H{
		"reset_token": resetToken,
		"email":       user.Email,
	})
}

// ResetPassword godoc
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if user == nil || user.PasswordResetCodeHash == nil || user.PasswordResetExpiresAt == nil {
		response.Error(c, http.StatusBadRequest, "Token reset tidak valid atau sudah kadaluarsa.", nil)
		return
	}
	if time.Now().After(*user.PasswordResetExpiresAt) {
		response.Error(c, http.StatusBadRequest, "Token reset tidak valid atau sudah kadaluarsa.", nil)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordResetCodeHash), []byte(req.ResetToken)); err != nil {
		response.Error(c, http.StatusBadRequest, "Token reset tidak valid atau sudah kadaluarsa.", nil)
		return
	}

	if msg := utils.ValidatePasswordStrength(req.NewPassword, user.FullName, user.Email); msg != "" {
		response.Error(c, http.StatusBadRequest, msg, nil)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memproses password.", nil)
		return
	}

	if err := h.userRepo.UpdatePassword(user.UID, string(hashed)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengubah password.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Password berhasil diubah. Silakan login.", nil)
}

// VerifyPhone godoc
// POST /api/v1/auth/verify-phone
func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	var req models.VerifyPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	valid, err := h.otpRepo.ConsumeValid(req.PhoneNumber, req.Code)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !valid {
		response.Error(c, http.StatusBadRequest, "Kode OTP tidak valid atau sudah kadaluarsa.", nil)
		return
	}

	user, err := h.userRepo.FindByPhone(req.PhoneNumber)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "Pengguna tidak ditemukan.", nil)
		return
	}

	if err := h.userRepo.SetPhoneVerified(user.UID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memverifikasi nomor telepon.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Nomor telepon berhasil diverifikasi. Silakan login.", nil)
}

// ResendOtp godoc
// POST /api/v1/auth/resend-otp
func (h *AuthHandler) ResendOtp(c *gin.Context) {
	var req models.ResendOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	if err := h.sendPhoneOtp(req.PhoneNumber); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengirim ulang kode OTP.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Kode OTP baru telah dikirim.", nil)
}

// GetProfile godoc
// GET /api/v1/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userUID := c.GetString("userUID")

	user, err := h.userRepo.FindByUID(userUID)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "Profil tidak ditemukan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Profil berhasil diambil.", gin.H{
		"uid":               user.UID,
		"full_name":         user.FullName,
		"email":             user.Email,
		"phone_number":      user.PhoneNumber,
		"is_verified_phone": user.IsVerifiedPhone,
		"is_approved_food":  user.IsApprovedFood,
		"is_approved_mart":  user.IsApprovedMart,
		"last_login":        user.LastLogin,
		"created_at":        user.CreatedAt,
	})
}


// UpdateFcmToken PATCH /api/v1/auth/fcm-token
func (h *AuthHandler) UpdateFcmToken(c *gin.Context) {
	userUID := c.GetString("userUID")

	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Token wajib diisi.", nil)
		return
	}
	if err := h.userRepo.UpdateFcmToken(userUID, req.Token); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan FCM token.", nil)
		return
	}
	response.Success(c, http.StatusOK, "FCM token berhasil disimpan.", nil)
}
