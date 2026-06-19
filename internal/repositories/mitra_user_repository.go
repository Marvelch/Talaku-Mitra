package repositories

import (
	"talaku_mitra/internal/models"
	"time"

	"gorm.io/gorm"
)

type MitraUserRepository struct {
	db *gorm.DB
}

func NewMitraUserRepository(db *gorm.DB) *MitraUserRepository {
	return &MitraUserRepository{db: db}
}

func (r *MitraUserRepository) FindByEmail(email string) (*models.MitraUser, error) {
	var user models.MitraUser
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *MitraUserRepository) FindByUID(uid string) (*models.MitraUser, error) {
	var user models.MitraUser
	err := r.db.Where("uid = ? AND deleted_at IS NULL", uid).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *MitraUserRepository) FindByPhone(phone string) (*models.MitraUser, error) {
	var user models.MitraUser
	err := r.db.Where("phone_number = ? AND deleted_at IS NULL", phone).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *MitraUserRepository) FindByRefreshToken(token string) (*models.MitraUser, error) {
	var user models.MitraUser
	err := r.db.Where("refresh_token = ? AND deleted_at IS NULL", token).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *MitraUserRepository) Create(user *models.MitraUser) error {
	return r.db.Create(user).Error
}

func (r *MitraUserRepository) SetRefreshToken(uid string, token *string) error {
	return r.db.Model(&models.MitraUser{}).Where("uid = ?", uid).Update("refresh_token", token).Error
}

func (r *MitraUserRepository) UpdateLastLogin(uid string) error {
	return r.db.Model(&models.MitraUser{}).Where("uid = ?", uid).
		Update("last_login", gorm.Expr("NOW()")).Error
}

func (r *MitraUserRepository) SetPasswordResetCode(uid string, hash *string, expiresAt *time.Time) error {
	return r.db.Model(&models.MitraUser{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"password_reset_code_hash":   hash,
		"password_reset_expires_at": expiresAt,
	}).Error
}

func (r *MitraUserRepository) UpdatePassword(uid, newPasswordHash string) error {
	return r.db.Model(&models.MitraUser{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"password_hash":              newPasswordHash,
		"password_reset_code_hash":   nil,
		"password_reset_expires_at": nil,
	}).Error
}

func (r *MitraUserRepository) SetPhoneVerified(uid string) error {
	return r.db.Model(&models.MitraUser{}).Where("uid = ?", uid).Update("is_verified_phone", true).Error
}

// FindAll returns all active mitra users (for backoffice listing).
func (r *MitraUserRepository) FindAll() ([]models.MitraUser, error) {
	var users []models.MitraUser
	err := r.db.Where("deleted_at IS NULL").Order("created_at DESC").Find(&users).Error
	return users, err
}

// HardDeleteByUID permanently removes an unverified mitra user record.
func (r *MitraUserRepository) HardDeleteByUID(uid string) error {
	return r.db.Unscoped().Where("uid = ?", uid).Delete(&models.MitraUser{}).Error
}

// UpdateCoreFields overwrites registration-related fields on an existing mitra user.
func (r *MitraUserRepository) UpdateCoreFields(uid string, updates map[string]interface{}) error {
	return r.db.Model(&models.MitraUser{}).Where("uid = ?", uid).Updates(updates).Error
}
