package repositories

import (
	"talaku_mitra/internal/models"

	"gorm.io/gorm"
)

type OtpRepository struct {
	db *gorm.DB
}

func NewOtpRepository(db *gorm.DB) *OtpRepository {
	return &OtpRepository{db: db}
}

func (r *OtpRepository) Create(otp *models.OtpVerification) error {
	return r.db.Create(otp).Error
}

func (r *OtpRepository) DeleteUnusedByPhone(phone string) error {
	return r.db.Where("phone = ? AND is_used = false", phone).Delete(&models.OtpVerification{}).Error
}

func (r *OtpRepository) DeleteUnusedByPhoneAndType(phone, otpType string) error {
	return r.db.Where("phone = ? AND type = ? AND is_used = false", phone, otpType).Delete(&models.OtpVerification{}).Error
}

// ConsumeValid atomically marks a matching, unused, unexpired OTP as used and
// reports whether such a record existed, avoiding read-then-write races.
func (r *OtpRepository) ConsumeValid(phone, code string) (bool, error) {
	result := r.db.Model(&models.OtpVerification{}).
		Where("phone = ? AND code = ? AND is_used = false AND expires_at > NOW()", phone, code).
		Update("is_used", true)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ConsumeValidByType is like ConsumeValid but also filters by OTP type.
func (r *OtpRepository) ConsumeValidByType(phone, code, otpType string) (bool, error) {
	result := r.db.Model(&models.OtpVerification{}).
		Where("phone = ? AND code = ? AND type = ? AND is_used = false AND expires_at > NOW()", phone, code, otpType).
		Update("is_used", true)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
