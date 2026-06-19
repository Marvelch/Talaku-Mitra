package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"talaku_mitra/internal/models"
	"talaku_mitra/internal/repositories"
	"talaku_mitra/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize    = 5 << 20 // 5 MB
	uploadsBaseDir = "./uploads"
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

type UploadHandler struct {
	uploadRepo *repositories.UploadRepository
	storeRepo  *repositories.StoreRepository
	foodRepo   *repositories.FoodRepository
}

func NewUploadHandler(
	uploadRepo *repositories.UploadRepository,
	storeRepo *repositories.StoreRepository,
	foodRepo *repositories.FoodRepository,
) *UploadHandler {
	return &UploadHandler{
		uploadRepo: uploadRepo,
		storeRepo:  storeRepo,
		foodRepo:   foodRepo,
	}
}

func validateImageFile(c *gin.Context, field string) (string, []byte, string, bool) {
	file, header, err := c.Request.FormFile(field)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File '"+field+"' wajib diupload.", nil)
		return "", nil, "", false
	}
	defer file.Close()

	if header.Size > maxFileSize {
		response.Error(c, http.StatusBadRequest, "Ukuran file maksimal 5 MB.", nil)
		return "", nil, "", false
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageExts[ext] {
		response.Error(c, http.StatusBadRequest, "Format file tidak didukung. Gunakan JPG, PNG, atau WEBP.", nil)
		return "", nil, "", false
	}

	buf := make([]byte, header.Size)
	if _, err := file.Read(buf); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membaca file.", nil)
		return "", nil, "", false
	}

	return header.Filename, buf, ext, true
}

func saveFile(dir, filename string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}
	// Return URL path (relative to server root)
	return "/" + filepath.ToSlash(filepath.Join(strings.TrimPrefix(dir, "./"), filename)), nil
}

// UploadStoreLogo godoc
// POST /api/v1/stores/:id/logo
func (h *UploadHandler) UploadStoreLogo(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	storeID := c.Param("id")

	isOwner, err := h.storeRepo.IsOwner(storeID, ownerUID)
	if err != nil || !isOwner {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk mengubah toko ini.", nil)
		return
	}

	_, data, ext, ok := validateImageFile(c, "logo")
	if !ok {
		return
	}

	filename := "logo" + ext
	dir := filepath.Join(uploadsBaseDir, "stores", storeID)
	url, err := saveFile(dir, filename, data)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan file.", nil)
		return
	}

	if err := h.uploadRepo.UpdateStoreLogoURL(storeID, url); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui logo toko.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Logo toko berhasil diupload.", gin.H{"logo_url": url})
}

// UploadStoreBanner godoc
// POST /api/v1/stores/:id/banner
func (h *UploadHandler) UploadStoreBanner(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	storeID := c.Param("id")

	isOwner, err := h.storeRepo.IsOwner(storeID, ownerUID)
	if err != nil || !isOwner {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk mengubah toko ini.", nil)
		return
	}

	_, data, ext, ok := validateImageFile(c, "banner")
	if !ok {
		return
	}

	filename := "banner" + ext
	dir := filepath.Join(uploadsBaseDir, "stores", storeID)
	url, err := saveFile(dir, filename, data)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan file.", nil)
		return
	}

	if err := h.uploadRepo.UpdateStoreBannerURL(storeID, url); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui banner toko.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Banner toko berhasil diupload.", gin.H{"banner_url": url})
}

// UploadFoodImage godoc
// POST /api/v1/stores/:id/foods/:food_id/images
func (h *UploadHandler) UploadFoodImage(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	foodID := c.Param("food_id")

	// Pastikan makanan milik owner
	isOwned, err := h.foodRepo.IsOwnedByUser(foodID, ownerUID)
	if err != nil || !isOwned {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk mengubah makanan ini.", nil)
		return
	}

	// Cek batas maksimal 5 foto
	count, err := h.uploadRepo.CountFoodImages(foodID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if count >= int64(models.MaxFoodImages) {
		response.Error(c, http.StatusBadRequest,
			fmt.Sprintf("Maksimal %d foto per makanan. Hapus foto lama sebelum menambah yang baru.", models.MaxFoodImages),
			nil,
		)
		return
	}

	_, data, ext, ok := validateImageFile(c, "image")
	if !ok {
		return
	}

	filename := uuid.New().String() + ext
	dir := filepath.Join(uploadsBaseDir, "foods", foodID)
	url, err := saveFile(dir, filename, data)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan file.", nil)
		return
	}

	img, err := h.uploadRepo.CreateFoodImage(foodID, url, int(count))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan data foto.", nil)
		return
	}

	// Foto pertama otomatis jadi image_url utama
	if count == 0 {
		_ = h.uploadRepo.UpdateFoodMainImage(foodID, url)
	}

	response.Success(c, http.StatusCreated, "Foto makanan berhasil diupload.", gin.H{
		"image":           img,
		"total_images":    count + 1,
		"remaining_slots": int64(models.MaxFoodImages) - count - 1,
	})
}

// GetFoodImages godoc
// GET /api/v1/foods/:id/images
func (h *UploadHandler) GetFoodImages(c *gin.Context) {
	foodID := c.Param("id")

	food, err := h.foodRepo.FindByID(foodID)
	if err != nil || food == nil {
		response.Error(c, http.StatusNotFound, "Makanan tidak ditemukan.", nil)
		return
	}

	images, err := h.uploadRepo.FindFoodImages(foodID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil foto makanan.", nil)
		return
	}

	response.Success(c, http.StatusOK, "Foto makanan berhasil diambil.", gin.H{
		"food_id":         foodID,
		"images":          images,
		"total":           len(images),
		"remaining_slots": models.MaxFoodImages - len(images),
	})
}

// DeleteFoodImage godoc
// DELETE /api/v1/stores/:id/foods/:food_id/images/:image_id
func (h *UploadHandler) DeleteFoodImage(c *gin.Context) {
	ownerUID := c.GetString("userUID")
	foodID := c.Param("food_id")
	imageID := c.Param("image_id")

	isOwned, err := h.foodRepo.IsOwnedByUser(foodID, ownerUID)
	if err != nil || !isOwned {
		response.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk menghapus foto ini.", nil)
		return
	}

	img, err := h.uploadRepo.FindFoodImageByID(imageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if img == nil || img.FoodID != foodID {
		response.Error(c, http.StatusNotFound, "Foto tidak ditemukan.", nil)
		return
	}

	// Hapus file fisik
	localPath := "." + img.URL
	_ = os.Remove(localPath)

	if err := h.uploadRepo.DeleteFoodImage(imageID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus foto.", nil)
		return
	}

	// Jika foto yang dihapus adalah image_url utama, update ke foto pertama yang tersisa
	food, _ := h.foodRepo.FindByID(foodID)
	if food != nil && food.ImageURL != nil && *food.ImageURL == img.URL {
		remaining, _ := h.uploadRepo.FindFoodImages(foodID)
		if len(remaining) > 0 {
			_ = h.uploadRepo.UpdateFoodMainImage(foodID, remaining[0].URL)
		} else {
			_ = h.uploadRepo.UpdateFoodMainImage(foodID, "")
		}
	}

	response.Success(c, http.StatusOK, "Foto berhasil dihapus.", nil)
}
