package models

import "time"

type FoodStatus string

const (
	FoodStatusAvailable   FoodStatus = "available"
	FoodStatusUnavailable FoodStatus = "unavailable"
)

type Food struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoreID     string     `gorm:"column:store_id;type:uuid;not null;index" json:"store_id"`
	Name        string     `gorm:"column:name;type:varchar(150);not null" json:"name"`
	Description *string    `gorm:"column:description;type:text" json:"description"`
	Price       float64    `gorm:"column:price;type:numeric(12,2);not null" json:"price"`
	Category    *string    `gorm:"column:category;type:varchar(50)" json:"category"`
	ImageURL    *string    `gorm:"column:image_url;type:varchar(255)" json:"image_url"`
	Status      FoodStatus `gorm:"column:status;type:varchar(20);default:'available'" json:"status"`
	IsRecommend bool       `gorm:"column:is_recommend;default:false" json:"is_recommend"`
	Stock       *int       `gorm:"column:stock;type:int" json:"stock"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index" json:"-"`

	Store *Store `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

type CreateFoodRequest struct {
	Name        string     `json:"name" binding:"required,min=2,max=150"`
	Description *string    `json:"description"`
	Price       float64    `json:"price" binding:"required,gt=0"`
	Category    *string    `json:"category"`
	IsRecommend bool       `json:"is_recommend"`
	Stock       *int       `json:"stock"`
	Status      FoodStatus `json:"status"`
}

type UpdateFoodRequest struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Price       *float64    `json:"price"`
	Category    *string     `json:"category"`
	IsRecommend *bool       `json:"is_recommend"`
	Stock       *int        `json:"stock"`
	Status      *FoodStatus `json:"status"`
}

func (Food) TableName() string { return "mitra_foods" }
