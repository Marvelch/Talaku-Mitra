package models

import "time"

type FoodOrderStatus string

const (
	FoodOrderWaitingDriver     FoodOrderStatus = "waiting_driver"
	FoodOrderWaitingRestaurant FoodOrderStatus = "waiting_restaurant"
	FoodOrderPreparing         FoodOrderStatus = "preparing"
	FoodOrderReady             FoodOrderStatus = "ready"
	FoodOrderOnDelivery        FoodOrderStatus = "on_delivery"
	FoodOrderDelivered         FoodOrderStatus = "delivered"
	FoodOrderCancelled         FoodOrderStatus = "cancelled"
)

type FoodOrder struct {
	ID      string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID  string          `gorm:"column:user_id;type:uuid;not null;index"                  json:"user_id"`
	DriverID *string        `gorm:"column:driver_id;type:uuid;index"                         json:"driver_id"`
	StoreID string          `gorm:"column:store_id;type:uuid;not null;index"                 json:"store_id"`
	Status  FoodOrderStatus `gorm:"column:status;type:varchar(30);not null;default:'waiting_driver';index" json:"status"`

	Subtotal    float64 `gorm:"column:subtotal;type:numeric(12,2);not null;default:0"    json:"subtotal"`
	DeliveryFee float64 `gorm:"column:delivery_fee;type:numeric(12,2);not null;default:0" json:"delivery_fee"`
	ServiceFee  float64 `gorm:"column:service_fee;type:numeric(12,2);not null;default:0" json:"service_fee"`
	Total       float64 `gorm:"column:total;type:numeric(12,2);not null;default:0"       json:"total"`

	DeliveryAddress string   `gorm:"column:delivery_address;type:text;not null" json:"delivery_address"`
	DeliveryLat     *float64 `gorm:"column:delivery_lat"                        json:"delivery_lat"`
	DeliveryLng     *float64 `gorm:"column:delivery_lng"                        json:"delivery_lng"`
	Note            string   `gorm:"column:note;type:text"                      json:"note"`

	VehicleTypeID   *int    `gorm:"column:vehicle_type_id"                    json:"vehicle_type_id"`
	VehicleTypeName *string `gorm:"column:vehicle_type_name;type:varchar(50)" json:"vehicle_type_name"`

	DriverAmount *float64 `gorm:"column:driver_amount;type:numeric(12,2)" json:"driver_amount"`
	TalakuGross  *float64 `gorm:"column:talaku_gross;type:numeric(12,2)"  json:"talaku_gross"`
	TaxAmount    *float64 `gorm:"column:tax_amount;type:numeric(12,2)"    json:"tax_amount"`
	TalakuNet    *float64 `gorm:"column:talaku_net;type:numeric(12,2)"    json:"talaku_net"`

	AcceptedAt  *time.Time `gorm:"column:accepted_at"   json:"accepted_at"`
	ConfirmedAt *time.Time `gorm:"column:confirmed_at"  json:"confirmed_at"`
	DeliveredAt *time.Time `gorm:"column:delivered_at"  json:"delivered_at"`
	CancelledAt *time.Time `gorm:"column:cancelled_at"  json:"cancelled_at"`
	CancelReason string    `gorm:"column:cancel_reason;type:text" json:"cancel_reason"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relasi (preload)
	Store *Store           `gorm:"foreignKey:StoreID;references:ID" json:"store,omitempty"`
	Items []FoodOrderItem  `gorm:"foreignKey:OrderID;references:ID" json:"items,omitempty"`
}

func (FoodOrder) TableName() string { return "food_orders" }

type FoodOrderItem struct {
	ID        string  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID   string  `gorm:"column:order_id;type:uuid;not null;index"                 json:"order_id"`
	FoodID    string  `gorm:"column:food_id;type:uuid;not null"                        json:"food_id"`
	FoodName  string  `gorm:"column:food_name;type:varchar(150);not null"              json:"food_name"`
	FoodPrice float64 `gorm:"column:food_price;type:numeric(12,2);not null"            json:"food_price"`
	Quantity  int     `gorm:"column:quantity;not null;default:1"                       json:"quantity"`
	Subtotal  float64 `gorm:"column:subtotal;type:numeric(12,2);not null"              json:"subtotal"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (FoodOrderItem) TableName() string { return "food_order_items" }

// ── Request/Response DTOs ──────────────────────────────────────────────────

type CreateFoodOrderItemInput struct {
	FoodID   string `json:"food_id"  binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

type CreateFoodOrderInput struct {
	StoreID         string                     `json:"store_id"         binding:"required"`
	Items           []CreateFoodOrderItemInput `json:"items"            binding:"required,min=1"`
	DeliveryAddress string                     `json:"delivery_address" binding:"required"`
	DeliveryLat     *float64                   `json:"delivery_lat"`
	DeliveryLng     *float64                   `json:"delivery_lng"`
	Note            string                     `json:"note"`
	VehicleTypeID   *int                       `json:"vehicle_type_id"`
	VehicleTypeName *string                    `json:"vehicle_type_name"`
	DeliveryFee     float64                    `json:"delivery_fee"`
	ServiceFee      float64                    `json:"service_fee"`
}

type FoodOrderResponse struct {
	ID              string          `json:"id"`
	Status          FoodOrderStatus `json:"status"`
	StoreName       string          `json:"store_name"`
	StorePhone      string          `json:"store_phone"`
	StoreAddress    string          `json:"store_address"`
	StoreLogoURL    string          `json:"store_logo_url"`
	Items           []FoodOrderItem `json:"items"`
	Subtotal        float64         `json:"subtotal"`
	DeliveryFee     float64         `json:"delivery_fee"`
	ServiceFee      float64         `json:"service_fee"`
	Total           float64         `json:"total"`
	DeliveryAddress string          `json:"delivery_address"`
	Note            string          `json:"note"`
	VehicleTypeName string          `json:"vehicle_type_name"`
	DriverID        *string         `json:"driver_id"`
	CancelReason    string          `json:"cancel_reason"`
	CreatedAt       string          `json:"created_at"`
}

// DTO untuk driver melihat daftar food order yang tersedia
type DriverFoodOrderItem struct {
	ID              string          `json:"id"`
	Status          FoodOrderStatus `json:"status"`
	StoreName       string          `json:"store_name"`
	StoreAddress    string          `json:"store_address"`
	StorePhone      string          `json:"store_phone"`
	StoreLogoURL    string          `json:"store_logo_url"`
	CustomerName    string          `json:"customer_name"`
	CustomerPhone   string          `json:"customer_phone"`
	DeliveryAddress string          `json:"delivery_address"`
	DeliveryLat     *float64        `json:"delivery_lat"`
	DeliveryLng     *float64        `json:"delivery_lng"`
	Items           []FoodOrderItem `json:"items"`
	Total           float64         `json:"total"`
	DeliveryFee     float64         `json:"delivery_fee"`
	DriverAmount    *float64        `json:"driver_amount"`
	VehicleTypeName string          `json:"vehicle_type_name"`
	CreatedAt       string          `json:"created_at"`
}
