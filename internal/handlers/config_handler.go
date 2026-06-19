package handlers

import (
	"net/http"
	"talaku_mitra/internal/config"
	"talaku_mitra/internal/repositories"
	"talaku_mitra/pkg/response"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	cfgRepo  *repositories.ConfigRepository
	userRepo *repositories.MitraUserRepository
}

func NewConfigHandler(cfgRepo *repositories.ConfigRepository, userRepo *repositories.MitraUserRepository) *ConfigHandler {
	return &ConfigHandler{cfgRepo: cfgRepo, userRepo: userRepo}
}

// isServiceEnabled reads a boolean config from app_configs; defaults to true if not found.
func (h *ConfigHandler) isServiceEnabled(key string) bool {
	cfg, err := h.cfgRepo.FindByKey(key)
	if err != nil || cfg == nil {
		return true // default: aktif
	}
	return cfg.ParameterValue == "true"
}

// GetServiceStatus godoc
// GET /api/v1/public/services
// Endpoint publik untuk membaca status layanan food & mart.
func (h *ConfigHandler) GetServiceStatus(c *gin.Context) {
	foodEnabled := h.isServiceEnabled("SERVICE_MAKANAN_ENABLED")
	martEnabled := h.isServiceEnabled("SERVICE_MART_ENABLED")

	response.Success(c, http.StatusOK, "Status layanan berhasil diambil.", gin.H{
		"food_enabled": foodEnabled,
		"mart_enabled": martEnabled,
	})
}

// BackofficeAuthMiddleware validates the X-Backoffice-Secret header.
func BackofficeAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Backoffice-Secret")
		if secret == "" || secret != config.AppConfig.BackofficeSecret {
			response.Error(c, http.StatusUnauthorized, "Akses tidak diizinkan.", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// ListMitra godoc
// GET /api/v1/backoffice/mitra
func (h *ConfigHandler) ListMitra(c *gin.Context) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil data mitra.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Data mitra berhasil diambil.", users)
}

// ToggleFoodApproval godoc
// PATCH /api/v1/backoffice/mitra/:uid/toggle-food
func (h *ConfigHandler) ToggleFoodApproval(c *gin.Context) {
	uid := c.Param("uid")
	user, err := h.userRepo.FindByUID(uid)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "Mitra tidak ditemukan.", nil)
		return
	}

	newVal := true
	if user.IsApprovedFood != nil {
		newVal = !*user.IsApprovedFood
	}
	if err := h.userRepo.UpdateCoreFields(uid, map[string]interface{}{"is_approved_food": newVal}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengubah status persetujuan.", nil)
		return
	}

	msg := "Mitra disetujui untuk layanan Food."
	if !newVal {
		msg = "Persetujuan layanan Food dicabut."
	}
	response.Success(c, http.StatusOK, msg, gin.H{"is_approved_food": newVal})
}

// ToggleMartApproval godoc
// PATCH /api/v1/backoffice/mitra/:uid/toggle-mart
func (h *ConfigHandler) ToggleMartApproval(c *gin.Context) {
	uid := c.Param("uid")
	user, err := h.userRepo.FindByUID(uid)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "Mitra tidak ditemukan.", nil)
		return
	}

	newVal := true
	if user.IsApprovedMart != nil {
		newVal = !*user.IsApprovedMart
	}
	if err := h.userRepo.UpdateCoreFields(uid, map[string]interface{}{"is_approved_mart": newVal}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengubah status persetujuan.", nil)
		return
	}

	msg := "Mitra disetujui untuk layanan Mart."
	if !newVal {
		msg = "Persetujuan layanan Mart dicabut."
	}
	response.Success(c, http.StatusOK, msg, gin.H{"is_approved_mart": newVal})
}

// UpdateMitraStatus godoc
// PATCH /api/v1/backoffice/mitra/:uid/status
// Body: {"is_active": true/false}
func (h *ConfigHandler) UpdateMitraStatus(c *gin.Context) {
	uid := c.Param("uid")
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid.", nil)
		return
	}
	if err := h.userRepo.UpdateCoreFields(uid, map[string]interface{}{"is_active": req.IsActive}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengubah status.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Status mitra berhasil diperbarui.", nil)
}

// ApprovalRequest for bulk approval
type ApprovalRequest struct {
	ApproveFood *bool `json:"approve_food"`
	ApproveMart *bool `json:"approve_mart"`
}

// UpdateMitraApproval godoc
// PATCH /api/v1/backoffice/mitra/:uid/approval
// Body: {"approve_food": true, "approve_mart": false}
func (h *ConfigHandler) UpdateMitraApproval(c *gin.Context) {
	uid := c.Param("uid")
	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid.", nil)
		return
	}

	user, err := h.userRepo.FindByUID(uid)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "Mitra tidak ditemukan.", nil)
		return
	}

	updates := map[string]interface{}{}
	if req.ApproveFood != nil {
		updates["is_approved_food"] = *req.ApproveFood
	}
	if req.ApproveMart != nil {
		updates["is_approved_mart"] = *req.ApproveMart
	}
	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "Tidak ada field yang diubah.", nil)
		return
	}

	if err := h.userRepo.UpdateCoreFields(uid, updates); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengubah persetujuan.", nil)
		return
	}

	// Reload to return updated state
	user, _ = h.userRepo.FindByUID(uid)
	response.Success(c, http.StatusOK, "Persetujuan mitra berhasil diperbarui.", gin.H{
		"uid":              user.UID,
		"full_name":        user.FullName,
		"email":            user.Email,
		"is_approved_food": user.IsApprovedFood,
		"is_approved_mart": user.IsApprovedMart,
	})
}

// GetMitraDetail godoc
// GET /api/v1/backoffice/mitra/:uid
func (h *ConfigHandler) GetMitraDetail(c *gin.Context) {
	uid := c.Param("uid")
	user, err := h.userRepo.FindByUID(uid)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "Mitra tidak ditemukan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Detail mitra berhasil diambil.", gin.H{
		"uid":               user.UID,
		"full_name":         user.FullName,
		"email":             user.Email,
		"phone_number":      user.PhoneNumber,
		"is_verified_phone": user.IsVerifiedPhone,
		"is_approved_food":  user.IsApprovedFood,
		"is_approved_mart":  user.IsApprovedMart,
		"is_active":         user.IsActive,
		"last_login":        user.LastLogin,
		"created_at":        user.CreatedAt,
	})
}

