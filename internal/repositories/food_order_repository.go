package repositories

import (
	"talaku_mitra/internal/models"
	"time"

	"gorm.io/gorm"
)

type FoodOrderRepository struct {
	db *gorm.DB
}

func NewFoodOrderRepository(db *gorm.DB) *FoodOrderRepository {
	return &FoodOrderRepository{db: db}
}

func (r *FoodOrderRepository) Create(order *models.FoodOrder, items []models.FoodOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		return tx.Create(&items).Error
	})
}

func (r *FoodOrderRepository) FindByID(id string) (*models.FoodOrder, error) {
	var order models.FoodOrder
	err := r.db.
		Preload("Store").
		Preload("Items").
		Where("id = ?", id).
		First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (r *FoodOrderRepository) FindByIDAndUserID(id, userID string) (*models.FoodOrder, error) {
	var order models.FoodOrder
	err := r.db.
		Preload("Store").
		Preload("Items").
		Where("id = ? AND user_id = ?", id, userID).
		First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

func (r *FoodOrderRepository) FindByIDAndDriverID(id, driverID string) (*models.FoodOrder, error) {
	var order models.FoodOrder
	err := r.db.
		Preload("Store").
		Preload("Items").
		Where("id = ? AND driver_id = ?", id, driverID).
		First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

// FindAvailableForDriver mengembalikan order makanan dengan status waiting_driver.
func (r *FoodOrderRepository) FindAvailableForDriver() ([]models.FoodOrder, error) {
	var orders []models.FoodOrder
	err := r.db.
		Preload("Store").
		Preload("Items").
		Where("status = ?", models.FoodOrderWaitingDriver).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// FindByStoreAndStatus mengembalikan order untuk restoran berdasarkan status.
func (r *FoodOrderRepository) FindByStoreAndStatus(storeID string, statuses []models.FoodOrderStatus) ([]models.FoodOrder, error) {
	var orders []models.FoodOrder
	err := r.db.
		Preload("Items").
		Where("store_id = ? AND status IN ?", storeID, statuses).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *FoodOrderRepository) UpdateStatus(id string, status models.FoodOrderStatus, extra map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	for k, v := range extra {
		updates[k] = v
	}
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", id).Updates(updates).Error
}

func (r *FoodOrderRepository) AssignDriver(orderID, driverID string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"driver_id":   driverID,
		"status":      models.FoodOrderPreparing,
		"accepted_at": now,
		"updated_at":  now,
	}).Error
}

func (r *FoodOrderRepository) ConfirmByRestaurant(orderID string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":       models.FoodOrderWaitingDriver,
		"confirmed_at": now,
		"updated_at":   now,
	}).Error
}

func (r *FoodOrderRepository) RejectByRestaurant(orderID, reason string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":        models.FoodOrderCancelled,
		"cancel_reason": reason,
		"cancelled_at":  now,
		"updated_at":    now,
	}).Error
}

func (r *FoodOrderRepository) MarkOnDelivery(orderID string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":     models.FoodOrderOnDelivery,
		"updated_at": now,
	}).Error
}

func (r *FoodOrderRepository) MarkDelivered(orderID string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":       models.FoodOrderDelivered,
		"delivered_at": now,
		"updated_at":   now,
	}).Error
}

func (r *FoodOrderRepository) MarkReady(orderID string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":     models.FoodOrderReady,
		"updated_at": now,
	}).Error
}

func (r *FoodOrderRepository) Cancel(orderID, reason string) error {
	now := time.Now()
	return r.db.Model(&models.FoodOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":        models.FoodOrderCancelled,
		"cancel_reason": reason,
		"cancelled_at":  now,
		"updated_at":    now,
	}).Error
}
