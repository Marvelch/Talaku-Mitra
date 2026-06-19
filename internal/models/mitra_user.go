package models

import "time"

// MitraUser adalah akun khusus mitra (merchant) food Talaku.
// Disimpan di tabel `mitra_users` terpisah dari tabel `users` yang dipakai
// oleh aplikasi Talaku utama (pelanggan, driver, dsb.).
type MitraUser struct {
	UID                    string     `gorm:"column:uid;type:uuid;primaryKey;default:gen_random_uuid()" json:"uid"`
	FullName               string     `gorm:"column:full_name;type:varchar(255);not null" json:"full_name"`
	Email                  string     `gorm:"column:email;type:varchar(255);not null" json:"email"`
	PasswordHash           string     `gorm:"column:password_hash;type:text;not null" json:"-"`
	PhoneNumber            *string    `gorm:"column:phone_number;type:varchar(20)" json:"phone_number"`
	IsVerifiedPhone        *bool      `gorm:"column:is_verified_phone;default:false" json:"is_verified_phone"`
	RefreshToken           *string    `gorm:"column:refresh_token;type:text" json:"-"`
	PasswordResetCodeHash  *string    `gorm:"column:password_reset_code_hash;type:text" json:"-"`
	PasswordResetExpiresAt *time.Time `gorm:"column:password_reset_expires_at" json:"-"`
	IsApprovedFood         *bool      `gorm:"column:is_approved_food;default:false" json:"is_approved_food"`
	IsApprovedMart         *bool      `gorm:"column:is_approved_mart;default:false" json:"is_approved_mart"`
	LastLogin              *time.Time `gorm:"column:last_login" json:"last_login"`
	IsActive               bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt              time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt              *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

func (MitraUser) TableName() string { return "mitra_users" }

// ── Request / Response types ──────────────────────────────────────────────────

type RegisterRequest struct {
	FullName    string `json:"full_name" binding:"required,min=2,max=255"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	// Identifier bisa berupa email atau nomor telepon (08xx / +62xx)
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
	// Method: "whatsapp" (default) atau "email"
	Method string `json:"method"`
}

type VerifyForgotOtpRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	ResetToken  string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
