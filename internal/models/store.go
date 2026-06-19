package models

import "time"

type StoreStatus string

const (
	StoreStatusActive   StoreStatus = "active"
	StoreStatusInactive StoreStatus = "inactive"
	StoreStatusClosed   StoreStatus = "closed"
)

type Store struct {
	ID          string      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerUID    string      `gorm:"column:owner_uid;type:uuid;not null;index" json:"owner_uid"`
	Name        string      `gorm:"column:name;type:varchar(150);not null" json:"name"`
	Description *string     `gorm:"column:description;type:text" json:"description"`
	Address     string      `gorm:"column:address;type:text;not null" json:"address"`
	Phone       *string     `gorm:"column:phone;type:varchar(20)" json:"phone"`
	LogoURL     *string     `gorm:"column:logo_url;type:varchar(255)" json:"logo_url"`
	BannerURL   *string     `gorm:"column:banner_url;type:varchar(255)" json:"banner_url"`
	Status      StoreStatus `gorm:"column:status;type:varchar(20);default:'active'" json:"status"`
	OpenTime    *string     `gorm:"column:open_time;type:varchar(10)" json:"open_time"`
	CloseTime   *string     `gorm:"column:close_time;type:varchar(10)" json:"close_time"`
	Latitude    *float64    `gorm:"column:latitude;type:double precision" json:"latitude"`
	Longitude   *float64    `gorm:"column:longitude;type:double precision" json:"longitude"`
	Rating      float64     `gorm:"column:rating;type:numeric(3,2);not null;default:0" json:"rating"`
	RatingCount int         `gorm:"column:rating_count;not null;default:0" json:"rating_count"`
	CreatedAt   time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time  `gorm:"column:deleted_at;index" json:"-"`

	Owner *MitraUser `gorm:"foreignKey:OwnerUID;references:UID" json:"owner,omitempty"`
	Foods []*Food `gorm:"foreignKey:StoreID" json:"foods,omitempty"`
}

type CreateStoreRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=150"`
	Description *string  `json:"description"`
	Address     string   `json:"address" binding:"required"`
	Phone       *string  `json:"phone"`
	OpenTime    *string  `json:"open_time"`
	CloseTime   *string  `json:"close_time"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type UpdateStoreRequest struct {
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	Address     *string      `json:"address"`
	Phone       *string      `json:"phone"`
	Status      *StoreStatus `json:"status"`
	OpenTime    *string      `json:"open_time"`
	CloseTime   *string      `json:"close_time"`
	Latitude    *float64     `json:"latitude"`
	Longitude   *float64     `json:"longitude"`
}

func (Store) TableName() string { return "mitra_stores" }

// StoreWithDistance embeds Store and includes the calculated distance from the customer.
type StoreWithDistance struct {
	Store
	DistanceKm float64 `json:"distance_km" gorm:"column:distance_km"`
}
