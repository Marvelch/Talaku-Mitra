package handlers

import (
	"net/http"
	"strconv"
	"talaku_mitra/internal/models"
	"talaku_mitra/internal/repositories"
	"talaku_mitra/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FoodHandler struct {
	foodRepo  *repositories.FoodRepository
	storeRepo *repositories.StoreRepository
}

func NewFoodHandler(foodRepo *repositories.FoodRepository, storeRepo *repositories.StoreRepository) *FoodHandler {
	return &FoodHandler{foodRepo: foodRepo, storeRepo: storeRepo}
}

// CreateFood godoc
// POST /api/v1/stores/:id/foods
func (h *FoodHandler) CreateFood(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	storeID := c.Param("id")

	isOwner, err := h.storeRepo.IsOwner(storeID, ownerUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !isOwner {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk menambah makanan di toko ini.", nil)
		return
	}

	var req models.CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	status := req.Status
	if status == "" {
		status = models.FoodStatusAvailable
	}

	food := &models.Food{
		ID:          uuid.New().String(),
		StoreID:     storeID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		IsRecommend: req.IsRecommend,
		Stock:       req.Stock,
		Status:      status,
	}

	if err := h.foodRepo.Create(food); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menambah makanan.", nil)
		return
	}

	response.Success(c, http.StatusCreated, "Makanan berhasil ditambahkan.", food)
}

// GetStoreFoods godoc
// GET /api/v1/stores/:id/foods
func (h *FoodHandler) GetStoreFoods(c *gin.Context) {
	storeID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	store, err := h.storeRepo.FindByID(storeID)
	if err != nil || store == nil {
		response.Error(c, http.StatusNotFound, "Toko tidak ditemukan.", nil)
		return
	}

	foods, total, err := h.foodRepo.FindByStoreID(storeID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data makanan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Data makanan berhasil diambil.", gin.H{
		"store": store,
		"foods": foods,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetAllFoods godoc
// GET /api/v1/foods
func (h *FoodHandler) GetAllFoods(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var catPtr *string
	if category != "" {
		catPtr = &category
	}

	isRecommendStr := c.Query("is_recommend")
	var isRecommendPtr *bool
	if isRecommendStr == "true" {
		v := true
		isRecommendPtr = &v
	}

	foods, total, err := h.foodRepo.FindAll(page, limit, catPtr, isRecommendPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data makanan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Data makanan berhasil diambil.", gin.H{
		"foods": foods,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetRandomPicks godoc
// GET /api/v1/foods/random-picks?lat=&lng=&radius=&limit=
//
// Returns one random available food from each of the nearby stores with the
// best rating (rating DESC, distance ASC as tiebreaker). Replaces the old
// "is_recommend" flag based picks with a location + rating aware selection:
// only stores within `radius` km (same island) are considered, and stores
// with no available food are skipped.
func (h *FoodHandler) GetRandomPicks(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	if latStr == "" || lngStr == "" {
		response.Error(c, http.StatusBadRequest, "Parameter lat dan lng wajib diisi.", nil)
		return
	}

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil {
		response.Error(c, http.StatusBadRequest, "Koordinat lat/lng tidak valid.", nil)
		return
	}

	radius := 15.0
	if r := c.Query("radius"); r != "" {
		if rv, err := strconv.ParseFloat(r, 64); err == nil && rv > 0 {
			radius = rv
		}
	}

	limit := 6
	if l := c.Query("limit"); l != "" {
		if lv, err := strconv.Atoi(l); err == nil && lv > 0 && lv <= 20 {
			limit = lv
		}
	}

	stores, err := h.storeRepo.FindNearbyByRating(lat, lng, radius, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data restoran terdekat.", nil)
		return
	}

	picks := make([]gin.H, 0, len(stores))
	for _, store := range stores {
		food, err := h.foodRepo.FindRandomByStoreID(store.ID)
		if err != nil || food == nil {
			continue // skip stores with no available food
		}
		picks = append(picks, gin.H{
			"id":           food.ID,
			"store_id":     store.ID,
			"name":         food.Name,
			"description":  food.Description,
			"price":        food.Price,
			"category":     food.Category,
			"image_url":    food.ImageURL,
			"status":       food.Status,
			"store_name":   store.Name,
			"store_rating": store.Rating,
			"distance_km":  store.DistanceKm,
		})
	}

	response.Success(c, http.StatusOK, "Menu pilihan berhasil diambil.", gin.H{
		"foods": picks,
	})
}

// GetFood godoc
// GET /api/v1/foods/:id
func (h *FoodHandler) GetFood(c *gin.Context) {
	food, err := h.foodRepo.FindByID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if food == nil {
		response.Error(c, http.StatusNotFound, "Makanan tidak ditemukan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Detail makanan berhasil diambil.", food)
}

// UpdateFood godoc
// PUT /api/v1/stores/:id/foods/:food_id
func (h *FoodHandler) UpdateFood(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	foodID := c.Param("food_id")

	isOwned, err := h.foodRepo.IsOwnedByUser(foodID, ownerUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !isOwned {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk mengubah makanan ini.", nil)
		return
	}

	var req models.UpdateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.IsRecommend != nil {
		updates["is_recommend"] = *req.IsRecommend
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "Tidak ada data yang diubah.", nil)
		return
	}

	if err := h.foodRepo.Update(foodID, updates); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui makanan.", nil)
		return
	}

	food, _ := h.foodRepo.FindByID(foodID)
	response.Success(c, http.StatusOK, "Makanan berhasil diperbarui.", food)
}

// DeleteFood godoc
// DELETE /api/v1/stores/:id/foods/:food_id
func (h *FoodHandler) DeleteFood(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	foodID := c.Param("food_id")

	isOwned, err := h.foodRepo.IsOwnedByUser(foodID, ownerUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !isOwned {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk menghapus makanan ini.", nil)
		return
	}

	if err := h.foodRepo.Delete(foodID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus makanan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Makanan berhasil dihapus.", nil)
}
