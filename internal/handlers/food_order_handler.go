package handlers

import (
	"errors"
	"log"
	"net/http"
	"talaku_mitra/internal/models"
	"talaku_mitra/internal/repositories"
	"talaku_mitra/pkg/fcm"
	"talaku_mitra/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mitraUserFcmFinder interface {
	FindFcmTokenByUID(uid string) (*string, error)
}

type FoodOrderHandler struct {
	orderRepo *repositories.FoodOrderRepository
	foodRepo  *repositories.FoodRepository
	storeRepo *repositories.StoreRepository
	userRepo  mitraUserFcmFinder
	fcm       *fcm.Service
}

func NewFoodOrderHandler(
	orderRepo *repositories.FoodOrderRepository,
	foodRepo *repositories.FoodRepository,
	storeRepo *repositories.StoreRepository,
	userRepo mitraUserFcmFinder,
	fcmSvc *fcm.Service,
) *FoodOrderHandler {
	return &FoodOrderHandler{orderRepo: orderRepo, foodRepo: foodRepo, storeRepo: storeRepo, userRepo: userRepo, fcm: fcmSvc}
}

// sendMitraFcm mengirim notif FCM ke pemilik toko berdasarkan storeID.
func (h *FoodOrderHandler) sendMitraFcm(storeID, title, body string, data map[string]string) {
	store, err := h.storeRepo.FindByID(storeID)
	if err != nil || store == nil {
		return
	}
	token, err := h.userRepo.FindFcmTokenByUID(store.OwnerUID)
	if err != nil || token == nil {
		return
	}
	if err := h.fcm.Send(*token, title, body, data); err != nil {
		log.Printf("[FCM] gagal kirim ke mitra %s: %v", store.OwnerUID, err)
	}
}

// ── Customer endpoints ─────────────────────────────────────────────────────

// CreateFoodOrder POST /api/v1/customer/food-orders
func (h *FoodOrderHandler) CreateFoodOrder(c *gin.Context) {
	userID := c.GetString("customerUID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "Tidak terautentikasi.", nil)
		return
	}

	var req models.CreateFoodOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid: "+err.Error(), nil)
		return
	}

	store, err := h.storeRepo.FindByID(req.StoreID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if store == nil || store.Status != models.StoreStatusActive {
		response.Error(c, http.StatusUnprocessableEntity, "Toko tidak ditemukan atau sedang tutup.", nil)
		return
	}

	var subtotal float64
	orderItems := make([]models.FoodOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		food, err := h.foodRepo.FindByID(item.FoodID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
			return
		}
		if food == nil || food.Status != models.FoodStatusAvailable {
			response.Error(c, http.StatusUnprocessableEntity, "Menu tidak tersedia.", nil)
			return
		}
		if food.StoreID != req.StoreID {
			response.Error(c, http.StatusUnprocessableEntity, "Menu tidak ada di toko ini.", nil)
			return
		}
		itemSubtotal := food.Price * float64(item.Quantity)
		subtotal += itemSubtotal
		orderItems = append(orderItems, models.FoodOrderItem{
			ID:        uuid.New().String(),
			FoodID:    food.ID,
			FoodName:  food.Name,
			FoodPrice: food.Price,
			Quantity:  item.Quantity,
			Subtotal:  itemSubtotal,
		})
	}

	total := subtotal + req.DeliveryFee + req.ServiceFee

	order := &models.FoodOrder{
		ID:              uuid.New().String(),
		UserID:          userID,
		StoreID:         req.StoreID,
		Status:          models.FoodOrderWaitingRestaurant,
		Subtotal:        subtotal,
		DeliveryFee:     req.DeliveryFee,
		ServiceFee:      req.ServiceFee,
		Total:           total,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryLat:     req.DeliveryLat,
		DeliveryLng:     req.DeliveryLng,
		Note:            req.Note,
		VehicleTypeID:   req.VehicleTypeID,
		VehicleTypeName: req.VehicleTypeName,
	}

	if err := h.orderRepo.Create(order, orderItems); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat pesanan: "+err.Error(), nil)
		return
	}

	// Notify mitra bahwa ada pesanan baru masuk — driver belum dilibatkan
	go h.sendMitraFcm(req.StoreID,
		"Pesanan Makanan Baru! 🍽️",
		"Ada pesanan baru menunggu konfirmasi Anda.",
		map[string]string{"type": "new_food_order", "order_id": order.ID},
	)

	response.Success(c, http.StatusCreated, "Pesanan berhasil dibuat.", gin.H{
		"order_id": order.ID,
		"status":   order.Status,
	})
}

// GetFoodOrderStatus GET /api/v1/customer/food-orders/:id
func (h *FoodOrderHandler) GetFoodOrderStatus(c *gin.Context) {
	userID := c.GetString("customerUID")
	orderID := c.Param("id")

	order, err := h.orderRepo.FindByIDAndUserID(orderID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}

	storeName, storePhone, storeAddress, storeLogoURL := "", "", "", ""
	if order.Store != nil {
		storeName = order.Store.Name
		storeAddress = order.Store.Address
		if order.Store.Phone != nil {
			storePhone = *order.Store.Phone
		}
		if order.Store.LogoURL != nil {
			storeLogoURL = *order.Store.LogoURL
		}
	}
	vehicleTypeName := ""
	if order.VehicleTypeName != nil {
		vehicleTypeName = *order.VehicleTypeName
	}

	response.Success(c, http.StatusOK, "OK", models.FoodOrderResponse{
		ID:              order.ID,
		Status:          order.Status,
		StoreName:       storeName,
		StorePhone:      storePhone,
		StoreAddress:    storeAddress,
		StoreLogoURL:    storeLogoURL,
		Items:           order.Items,
		Subtotal:        order.Subtotal,
		DeliveryFee:     order.DeliveryFee,
		ServiceFee:      order.ServiceFee,
		Total:           order.Total,
		DeliveryAddress: order.DeliveryAddress,
		Note:            order.Note,
		VehicleTypeName: vehicleTypeName,
		DriverID:        order.DriverID,
		CancelReason:    order.CancelReason,
		CreatedAt:       order.CreatedAt.Format(time.RFC3339),
	})
}

// CancelFoodOrder PATCH /api/v1/customer/food-orders/:id/cancel
func (h *FoodOrderHandler) CancelFoodOrder(c *gin.Context) {
	userID := c.GetString("customerUID")
	orderID := c.Param("id")

	order, err := h.orderRepo.FindByIDAndUserID(orderID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderWaitingDriver && order.Status != models.FoodOrderWaitingRestaurant {
		response.Error(c, http.StatusUnprocessableEntity, "Pesanan tidak dapat dibatalkan pada status ini.", nil)
		return
	}
	if err := h.orderRepo.Cancel(orderID, "Dibatalkan oleh customer"); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membatalkan pesanan.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Pesanan berhasil dibatalkan.", nil)
}

// ── Driver endpoints ───────────────────────────────────────────────────────

// GetAvailableFoodOrders GET /api/v1/driver/food-orders
func (h *FoodOrderHandler) GetAvailableFoodOrders(c *gin.Context) {
	orders, err := h.orderRepo.FindAvailableForDriver()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}

	result := make([]models.DriverFoodOrderItem, 0, len(orders))
	for _, o := range orders {
		storeName, storeAddress, storePhone, storeLogoURL := "", "", "", ""
		if o.Store != nil {
			storeName = o.Store.Name
			storeAddress = o.Store.Address
			if o.Store.Phone != nil {
				storePhone = *o.Store.Phone
			}
			if o.Store.LogoURL != nil {
				storeLogoURL = *o.Store.LogoURL
			}
		}
		vehicleTypeName := ""
		if o.VehicleTypeName != nil {
			vehicleTypeName = *o.VehicleTypeName
		}
		result = append(result, models.DriverFoodOrderItem{
			ID:              o.ID,
			Status:          o.Status,
			StoreName:       storeName,
			StoreAddress:    storeAddress,
			StorePhone:      storePhone,
			StoreLogoURL:    storeLogoURL,
			DeliveryAddress: o.DeliveryAddress,
			DeliveryLat:     o.DeliveryLat,
			DeliveryLng:     o.DeliveryLng,
			Items:           o.Items,
			Total:           o.Total,
			DeliveryFee:     o.DeliveryFee,
			DriverAmount:    o.DriverAmount,
			VehicleTypeName: vehicleTypeName,
			CreatedAt:       o.CreatedAt.Format(time.RFC3339),
		})
	}
	response.Success(c, http.StatusOK, "OK", gin.H{"orders": result})
}

// AcceptFoodOrder PATCH /api/v1/driver/food-orders/:id/accept
func (h *FoodOrderHandler) AcceptFoodOrder(c *gin.Context) {
	driverID := c.GetString("driverUID")
	orderID := c.Param("id")

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderWaitingDriver {
		response.Error(c, http.StatusConflict, "Pesanan sudah diambil driver lain.", nil)
		return
	}
	if err := h.orderRepo.AssignDriver(orderID, driverID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menerima pesanan.", nil)
		return
	}

	order, _ = h.orderRepo.FindByID(orderID)
	storeName, storeAddress, storePhone := "", "", ""
	if order != nil && order.Store != nil {
		storeName = order.Store.Name
		storeAddress = order.Store.Address
		if order.Store.Phone != nil {
			storePhone = *order.Store.Phone
		}
	}

	// Notify mitra bahwa driver dikonfirmasi — pesanan bisa diproses
	if order != nil {
		go h.sendMitraFcm(order.StoreID,
			"Driver Dikonfirmasi! 🛵",
			"Pesanan bisa diproses/dibuat, driver sedang menuju restoran.",
			map[string]string{"type": "driver_accepted", "order_id": orderID},
		)
	}

	response.Success(c, http.StatusOK, "Pesanan berhasil diterima.", gin.H{
		"order_id":      orderID,
		"status":        models.FoodOrderPreparing,
		"store_name":    storeName,
		"store_address": storeAddress,
		"store_phone":   storePhone,
	})
}

// GetDriverFoodOrderDetail GET /api/v1/driver/food-orders/:id
func (h *FoodOrderHandler) GetDriverFoodOrderDetail(c *gin.Context) {
	driverID := c.GetString("driverUID")
	orderID := c.Param("id")

	order, err := h.orderRepo.FindByIDAndDriverID(orderID, driverID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}

	storeName, storeAddress, storePhone, storeLogoURL := "", "", "", ""
	if order.Store != nil {
		storeName = order.Store.Name
		storeAddress = order.Store.Address
		if order.Store.Phone != nil {
			storePhone = *order.Store.Phone
		}
		if order.Store.LogoURL != nil {
			storeLogoURL = *order.Store.LogoURL
		}
	}

	response.Success(c, http.StatusOK, "OK", gin.H{
		"id":               order.ID,
		"status":           order.Status,
		"store_name":       storeName,
		"store_address":    storeAddress,
		"store_phone":      storePhone,
		"store_logo_url":   storeLogoURL,
		"items":            order.Items,
		"total":            order.Total,
		"delivery_fee":     order.DeliveryFee,
		"delivery_address": order.DeliveryAddress,
		"delivery_lat":     order.DeliveryLat,
		"delivery_lng":     order.DeliveryLng,
		"note":             order.Note,
		"cancel_reason":    order.CancelReason,
		"created_at":       order.CreatedAt.Format(time.RFC3339),
	})
}

// MarkOnDelivery PATCH /api/v1/driver/food-orders/:id/pickup
func (h *FoodOrderHandler) MarkOnDelivery(c *gin.Context) {
	driverID := c.GetString("driverUID")
	orderID := c.Param("id")

	order, err := h.orderRepo.FindByIDAndDriverID(orderID, driverID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderPreparing && order.Status != models.FoodOrderReady {
		response.Error(c, http.StatusUnprocessableEntity, "Status pesanan tidak dapat diperbarui.", nil)
		return
	}
	if err := h.orderRepo.MarkOnDelivery(orderID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui status.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Pesanan sedang dalam pengiriman.", nil)
}

// MarkDelivered PATCH /api/v1/driver/food-orders/:id/delivered
func (h *FoodOrderHandler) MarkDelivered(c *gin.Context) {
	driverID := c.GetString("driverUID")
	orderID := c.Param("id")

	order, err := h.orderRepo.FindByIDAndDriverID(orderID, driverID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}
	if order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderOnDelivery {
		response.Error(c, http.StatusUnprocessableEntity, "Pesanan belum dalam status pengiriman.", nil)
		return
	}
	if err := h.orderRepo.MarkDelivered(orderID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyelesaikan pesanan.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Pesanan berhasil diantarkan.", nil)
}

// ── Mitra (Restaurant) endpoints ──────────────────────────────────────────

// GetMitraFoodOrders GET /api/v1/mitra/food-orders
func (h *FoodOrderHandler) GetMitraFoodOrders(c *gin.Context) {
	mitraUID := c.GetString("userUID")

	stores, err := h.storeRepo.FindByOwnerUID(mitraUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server.", nil)
		return
	}

	storeIDs := make([]string, 0, len(stores))
	for _, s := range stores {
		storeIDs = append(storeIDs, s.ID)
	}
	if len(storeIDs) == 0 {
		response.Success(c, http.StatusOK, "OK", gin.H{"orders": []interface{}{}})
		return
	}

	activeStatuses := []models.FoodOrderStatus{
		models.FoodOrderWaitingDriver,
		models.FoodOrderWaitingRestaurant,
		models.FoodOrderPreparing,
		models.FoodOrderReady,
	}

	var allOrders []models.FoodOrder
	for _, storeID := range storeIDs {
		orders, err := h.orderRepo.FindByStoreAndStatus(storeID, activeStatuses)
		if err != nil {
			continue
		}
		allOrders = append(allOrders, orders...)
	}

	response.Success(c, http.StatusOK, "OK", gin.H{"orders": allOrders})
}

// ConfirmFoodOrder PATCH /api/v1/mitra/food-orders/:id/confirm
func (h *FoodOrderHandler) ConfirmFoodOrder(c *gin.Context) {
	mitraUID := c.GetString("userUID")
	orderID := c.Param("id")

	if err := h.validateMitraOwnsOrder(mitraUID, orderID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), nil)
		return
	}

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil || order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderWaitingRestaurant {
		response.Error(c, http.StatusUnprocessableEntity, "Pesanan tidak dalam status menunggu konfirmasi.", nil)
		return
	}
	if err := h.orderRepo.ConfirmByRestaurant(orderID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengkonfirmasi pesanan.", nil)
		return
	}

	// Setelah restoran terima, broadcast ke semua driver online
	go func() {
		store, err := h.storeRepo.FindByID(order.StoreID)
		storeName := ""
		if err == nil && store != nil {
			storeName = store.Name
		}
		if err := h.fcm.SendToTopic(
			"food_order_available",
			"Pesanan Makanan Tersedia! 🛵",
			"Ada pesanan makanan dari "+storeName+" menunggu driver.",
			map[string]string{"type": "new_food_order", "order_id": orderID},
		); err != nil {
			log.Printf("[FCM] gagal kirim ke topic driver: %v", err)
		}
	}()

	response.Success(c, http.StatusOK, "Pesanan dikonfirmasi. Menunggu driver.", nil)
}

// RejectFoodOrder PATCH /api/v1/mitra/food-orders/:id/reject
func (h *FoodOrderHandler) RejectFoodOrder(c *gin.Context) {
	mitraUID := c.GetString("userUID")
	orderID := c.Param("id")

	if err := h.validateMitraOwnsOrder(mitraUID, orderID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), nil)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	reason := req.Reason
	if reason == "" {
		reason = "Pesanan ditolak oleh restoran"
	}

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil || order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderWaitingRestaurant {
		response.Error(c, http.StatusUnprocessableEntity, "Pesanan tidak dalam status menunggu konfirmasi.", nil)
		return
	}
	if err := h.orderRepo.RejectByRestaurant(orderID, reason); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menolak pesanan.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Pesanan ditolak.", nil)
}

// MarkFoodReady PATCH /api/v1/mitra/food-orders/:id/ready
func (h *FoodOrderHandler) MarkFoodReady(c *gin.Context) {
	mitraUID := c.GetString("userUID")
	orderID := c.Param("id")

	if err := h.validateMitraOwnsOrder(mitraUID, orderID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), nil)
		return
	}

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil || order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderPreparing {
		response.Error(c, http.StatusUnprocessableEntity, "Pesanan tidak dalam status sedang diproses.", nil)
		return
	}
	if err := h.orderRepo.MarkReady(orderID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui status pesanan.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Makanan siap untuk diambil driver.", nil)
}

// MitraCancelFoodOrder PATCH /api/v1/mitra/food-orders/:id/cancel
func (h *FoodOrderHandler) MitraCancelFoodOrder(c *gin.Context) {
	mitraUID := c.GetString("userUID")
	orderID := c.Param("id")

	if err := h.validateMitraOwnsOrder(mitraUID, orderID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error(), nil)
		return
	}

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil || order == nil {
		response.Error(c, http.StatusNotFound, "Pesanan tidak ditemukan.", nil)
		return
	}
	if order.Status != models.FoodOrderWaitingDriver && order.Status != models.FoodOrderWaitingRestaurant {
		response.Error(c, http.StatusUnprocessableEntity, "Pesanan tidak dapat dibatalkan pada status ini.", nil)
		return
	}
	if err := h.orderRepo.Cancel(orderID, "Dibatalkan oleh restoran"); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membatalkan pesanan.", nil)
		return
	}
	response.Success(c, http.StatusOK, "Pesanan berhasil dibatalkan.", nil)
}

func (h *FoodOrderHandler) validateMitraOwnsOrder(mitraUID, orderID string) error {
	order, err := h.orderRepo.FindByID(orderID)
	if err != nil || order == nil {
		return errors.New("pesanan tidak ditemukan")
	}
	isOwner, err := h.storeRepo.IsOwner(order.StoreID, mitraUID)
	if err != nil || !isOwner {
		return errors.New("anda tidak memiliki izin untuk order ini")
	}
	return nil
}
