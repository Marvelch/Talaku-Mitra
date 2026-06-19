package repositories

import (
	"talaku_mitra/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

func (r *UploadRepository) CountFoodImages(foodID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.FoodImage{}).Where("food_id = ?", foodID).Count(&count).Error
	return count, err
}

func (r *UploadRepository) CreateFoodImage(foodID, url string, sortOrder int) (*models.FoodImage, error) {
	img := &models.FoodImage{
		ID:        uuid.New().String(),
		FoodID:    foodID,
		URL:       url,
		SortOrder: sortOrder,
	}
	if err := r.db.Create(img).Error; err != nil {
		return nil, err
	}
	return img, nil
}

func (r *UploadRepository) FindFoodImageByID(id string) (*models.FoodImage, error) {
	var img models.FoodImage
	err := r.db.Where("id = ?", id).First(&img).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &img, err
}

func (r *UploadRepository) FindFoodImages(foodID string) ([]*models.FoodImage, error) {
	var images []*models.FoodImage
	err := r.db.Where("food_id = ?", foodID).Order("sort_order ASC, created_at ASC").Find(&images).Error
	return images, err
}

func (r *UploadRepository) DeleteFoodImage(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.FoodImage{}).Error
}

func (r *UploadRepository) UpdateStoreLogoURL(storeID, url string) error {
	return r.db.Table("mitra_stores").Where("id = ?", storeID).Update("logo_url", url).Error
}

func (r *UploadRepository) UpdateStoreBannerURL(storeID, url string) error {
	return r.db.Table("mitra_stores").Where("id = ?", storeID).Update("banner_url", url).Error
}

func (r *UploadRepository) UpdateFoodMainImage(foodID, url string) error {
	return r.db.Table("mitra_foods").Where("id = ?", foodID).Update("image_url", url).Error
}
