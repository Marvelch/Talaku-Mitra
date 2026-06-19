package models

// AppConfig membaca tabel app_configs yang dikelola oleh layanan utama Talaku.
// Service ini membaca nilai konfigurasi layanan food dan mart secara read-only.
type AppConfig struct {
	UID            string `gorm:"column:uid;type:uuid;primaryKey" json:"uid"`
	ParameterKey   string `gorm:"column:parameter_key;type:varchar(100)" json:"parameter_key"`
	ParameterValue string `gorm:"column:parameter_value;type:text" json:"parameter_value"`
	Description    string `gorm:"column:description;type:text" json:"description"`
	IsActive       bool   `gorm:"column:is_active;default:true" json:"is_active"`
}

func (AppConfig) TableName() string { return "app_configs" }
