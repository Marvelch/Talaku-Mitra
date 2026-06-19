package repositories

import (
	"talaku_mitra/internal/models"

	"gorm.io/gorm"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// FindByKey returns nil (not error) when the config key doesn't exist.
func (r *ConfigRepository) FindByKey(key string) (*models.AppConfig, error) {
	var cfg models.AppConfig
	err := r.db.Where("parameter_key = ? AND is_active = true", key).First(&cfg).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &cfg, err
}
