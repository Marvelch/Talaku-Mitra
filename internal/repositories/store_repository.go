package repositories

import (
	"talaku_mitra/internal/models"

	"gorm.io/gorm"
)

type StoreRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

func (r *StoreRepository) Create(store *models.Store) error {
	return r.db.Create(store).Error
}

func (r *StoreRepository) FindByID(id string) (*models.Store, error) {
	var store models.Store
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&store).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &store, err
}

func (r *StoreRepository) FindByOwnerUID(ownerUID string) ([]*models.Store, error) {
	var stores []*models.Store
	err := r.db.Where("owner_uid = ? AND deleted_at IS NULL", ownerUID).Find(&stores).Error
	return stores, err
}

func (r *StoreRepository) FindAll(page, limit int) ([]*models.Store, int64, error) {
	var stores []*models.Store
	var total int64
	offset := (page - 1) * limit

	base := r.db.Model(&models.Store{}).Where("deleted_at IS NULL AND status = 'active'")
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Find(&stores).Error
	return stores, total, err
}

// FindNearby returns active stores within radiusKm of (lat, lng), sorted by distance.
// Uses PostgreSQL Haversine formula computed in SQL so filtering happens server-side.
// Only stores with non-null coordinates are considered.
// Default radius recommendation: 15 km — tight enough to stay on the same island.
func (r *StoreRepository) FindNearby(lat, lng, radiusKm float64, page, limit int) ([]models.StoreWithDistance, int64, error) {
	offset := (page - 1) * limit

	// Haversine formula expressed as a SQL expression.
	// Params order: lat, lat, lng  (3 placeholders per usage)
	const haversineExpr = `6371.0 * 2.0 * asin(sqrt(
		power(sin(radians((latitude - ?) / 2.0)), 2.0) +
		cos(radians(?)) * cos(radians(latitude)) *
		power(sin(radians((longitude - ?) / 2.0)), 2.0)
	))`

	// CTE computes distance once, outer query filters + paginates.
	const cte = `
		WITH nearby AS (
			SELECT *,
				` + haversineExpr + ` AS distance_km
			FROM mitra_stores
			WHERE deleted_at IS NULL
			  AND status = 'active'
			  AND latitude IS NOT NULL
			  AND longitude IS NOT NULL
		)
	`

	var total int64
	countErr := r.db.Raw(
		cte+`SELECT COUNT(*) FROM nearby WHERE distance_km <= ?`,
		lat, lat, lng, radiusKm,
	).Scan(&total).Error
	if countErr != nil {
		return nil, 0, countErr
	}

	var stores []models.StoreWithDistance
	err := r.db.Raw(
		cte+`SELECT * FROM nearby WHERE distance_km <= ? ORDER BY distance_km ASC LIMIT ? OFFSET ?`,
		lat, lat, lng, radiusKm, limit, offset,
	).Scan(&stores).Error

	return stores, total, err
}

// FindNearbyByRating returns active stores within radiusKm of (lat, lng),
// sorted by best rating first (distance as tiebreaker). Used to pick which
// nearby stores should contribute a "menu pilihan" item.
func (r *StoreRepository) FindNearbyByRating(lat, lng, radiusKm float64, limit int) ([]models.StoreWithDistance, error) {
	const haversineExpr = `6371.0 * 2.0 * asin(sqrt(
		power(sin(radians((latitude - ?) / 2.0)), 2.0) +
		cos(radians(?)) * cos(radians(latitude)) *
		power(sin(radians((longitude - ?) / 2.0)), 2.0)
	))`

	const cte = `
		WITH nearby AS (
			SELECT *,
				` + haversineExpr + ` AS distance_km
			FROM mitra_stores
			WHERE deleted_at IS NULL
			  AND status = 'active'
			  AND latitude IS NOT NULL
			  AND longitude IS NOT NULL
		)
	`

	var stores []models.StoreWithDistance
	err := r.db.Raw(
		cte+`SELECT * FROM nearby WHERE distance_km <= ?
			ORDER BY rating DESC, distance_km ASC LIMIT ?`,
		lat, lat, lng, radiusKm, limit,
	).Scan(&stores).Error

	return stores, err
}

func (r *StoreRepository) Update(id string, updates map[string]interface{}) error {
	return r.db.Model(&models.Store{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (r *StoreRepository) Delete(id string) error {
	return r.db.Model(&models.Store{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *StoreRepository) IsOwner(storeID, ownerUID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Store{}).
		Where("id = ? AND owner_uid = ? AND deleted_at IS NULL", storeID, ownerUID).
		Count(&count).Error
	return count > 0, err
}
