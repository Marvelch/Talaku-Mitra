package models

import "time"

// OtpVerification uses a "mitra_" prefixed table name to avoid colliding with
// the unrelated otp_verifications table owned by other Talaku services sharing
// the same database, following the same convention as Store ("mitra_stores").
type OtpVerification struct {
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	Type      string    `gorm:"column:type;type:varchar(50);not null" json:"type"`
	Phone     string    `gorm:"column:phone;type:varchar(20);not null;index" json:"phone"`
	Code      string    `gorm:"column:code;type:varchar(10);not null" json:"-"`
	IsUsed    bool      `gorm:"column:is_used;default:false" json:"is_used"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OtpVerification) TableName() string { return "mitra_otp_verifications" }

type VerifyPhoneRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type ResendOtpRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}
