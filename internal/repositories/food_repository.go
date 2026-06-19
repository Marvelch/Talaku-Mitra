package repositories

import (
	"talaku_mitra/internal/models"

	"gorm.io/gorm"
)

type FoodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) *FoodRepository {
	return &FoodRepository{db: db}
}

func (r *FoodRepository) Create(food *models.Food) error {
	return r.db.Create(food).Error
}

func (r *FoodRepository) FindByID(id string) (*models.Food, error) {
	var food models.Food
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&food).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &food, err
}

func (r *FoodRepository) FindByStoreID(storeID string, page, limit int) ([]*models.Food, int64, error) {
	var foods []*models.Food
	var total int64
	offset := (page - 1) * limit

	base := r.db.Model(&models.Food{}).Where("store_id = ? AND deleted_at IS NULL", storeID)
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Find(&foods).Error
	return foods, total, err
}

func (r *FoodRepository) FindAll(page, limit int, category *string, isRecommend *bool) ([]*models.Food, int64, error) {
	var foods []*models.Food
	var total int64
	offset := (page - 1) * limit

	base := r.db.Model(&models.Food{}).Where("deleted_at IS NULL AND status = 'available'")
	if category != nil && *category != "" {
		base = base.Where("category = ?", *category)
	}
	if isRecommend != nil {
		base = base.Where("is_recommend = ?", *isRecommend)
	}

	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Preload("Store").Find(&foods).Error
	return foods, total, err
}

// FindRandomByStoreID returns one random available food from the given store.
// Used to build "menu pilihan" picks from nearby stores sorted by rating.
func (r *FoodRepository) FindRandomByStoreID(storeID string) (*models.Food, error) {
	var food models.Food
	err := r.db.
		Where("store_id = ? AND deleted_at IS NULL AND status = 'available'", storeID).
		Order("random()").
		Limit(1).
		First(&food).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &food, err
}

func (r *FoodRepository) Update(id string, updates map[string]interface{}) error {
	return r.db.Model(&models.Food{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (r *FoodRepository) Delete(id string) error {
	return r.db.Model(&models.Food{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *FoodRepository) IsOwnedByUser(foodID, ownerUID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Food{}).
		Joins("JOIN mitra_stores ON mitra_foods.store_id = mitra_stores.id").
		Where("mitra_foods.id = ? AND mitra_stores.owner_uid = ? AND mitra_foods.deleted_at IS NULL AND mitra_stores.deleted_at IS NULL", foodID, ownerUID).
		Count(&count).Error
	return count > 0, err
}
