package models

import "time"

const MaxFoodImages = 5

type FoodImage struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FoodID    string    `gorm:"column:food_id;type:uuid;not null;index" json:"food_id"`
	URL       string    `gorm:"column:url;type:varchar(255);not null" json:"url"`
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (FoodImage) TableName() string { return "mitra_food_images" }
