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

type StoreHandler struct {
	storeRepo *repositories.StoreRepository
}

func NewStoreHandler(storeRepo *repositories.StoreRepository) *StoreHandler {
	return &StoreHandler{storeRepo: storeRepo}
}

// CreateStore godoc
// POST /api/v1/stores
func (h *StoreHandler) CreateStore(c *gin.Context) {
	ownerUID := c.GetString("userUID")

	var req models.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	store := &models.Store{
		ID:          uuid.New().String(),
		OwnerUID:    ownerUID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		OpenTime:    req.OpenTime,
		CloseTime:   req.CloseTime,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      models.StoreStatusActive,
	}

	if err := h.storeRepo.Create(store); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat toko.", nil)
		return
	}

	response.Success(c, http.StatusCreated, "Toko berhasil dibuat.", store)
}

// GetMyStores godoc
// GET /api/v1/stores/my
func (h *StoreHandler) GetMyStores(c *gin.Context) {
	ownerUID := c.GetString("userUID")

	stores, err := h.storeRepo.FindByOwnerUID(ownerUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data toko.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Data toko berhasil diambil.", stores)
}

// GetStores godoc
// GET /api/v1/stores
//
// Query params:
//
//	page, limit        — pagination (default 1, 10)
//	lat, lng           — customer coordinates (float64); when both are present the
//	                     response contains only stores within `radius` km and each
//	                     store object includes a `distance_km` field.
//	radius             — filter radius in km (default 15.0); use a tight value so
//	                     that stores on other islands (separated by water) are excluded.
func (h *StoreHandler) GetStores(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	if latStr != "" && lngStr != "" {
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

		stores, total, err := h.storeRepo.FindNearby(lat, lng, radius, page, limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Gagal mengambil data toko terdekat.", nil)
			return
		}

		response.Success(c, http.StatusOK, "Data toko terdekat berhasil diambil.", gin.H{
			"stores": stores,
			"pagination": gin.H{
				"page":   page,
				"limit":  limit,
				"total":  total,
				"lat":    lat,
				"lng":    lng,
				"radius": radius,
			},
		})
		return
	}

	stores, total, err := h.storeRepo.FindAll(page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data toko.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Data toko berhasil diambil.", gin.H{
		"stores": stores,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetStore godoc
// GET /api/v1/stores/:id
func (h *StoreHandler) GetStore(c *gin.Context) {
	store, err := h.storeRepo.FindByID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if store == nil {
		response.Error(c, http.StatusNotFound, "Toko tidak ditemukan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Detail toko berhasil diambil.", store)
}

// UpdateStore godoc
// PUT /api/v1/stores/:id
func (h *StoreHandler) UpdateStore(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	storeID := c.Param("id")

	isOwner, err := h.storeRepo.IsOwner(storeID, ownerUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !isOwner {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk mengubah toko ini.", nil)
		return
	}

	var req models.UpdateStoreRequest
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
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.OpenTime != nil {
		updates["open_time"] = *req.OpenTime
	}
	if req.CloseTime != nil {
		updates["close_time"] = *req.CloseTime
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "Tidak ada data yang diubah.", nil)
		return
	}

	if err := h.storeRepo.Update(storeID, updates); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui toko.", nil)
		return
	}

	store, _ := h.storeRepo.FindByID(storeID)
	response.Success(c, http.StatusOK, "Toko berhasil diperbarui.", store)
}

// DeleteStore godoc
// DELETE /api/v1/stores/:id
func (h *StoreHandler) DeleteStore(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	storeID := c.Param("id")

	isOwner, err := h.storeRepo.IsOwner(storeID, ownerUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if !isOwner {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk menghapus toko ini.", nil)
		return
	}

	if err := h.storeRepo.Delete(storeID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus toko.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Toko berhasil dihapus.", nil)
}
