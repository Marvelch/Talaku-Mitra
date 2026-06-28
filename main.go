package main

import (
	"log"
	"talaku_mitra/internal/config"
	"talaku_mitra/internal/handlers"
	"talaku_mitra/internal/models"
	"talaku_mitra/internal/repositories"
	"talaku_mitra/internal/routes"
	"talaku_mitra/pkg/fcm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()
	config.ConnectDB()

	if err := config.DB.AutoMigrate(
		&models.MitraUser{},
		&models.Store{},
		&models.Food{},
		&models.FoodImage{},
		&models.OtpVerification{},
		&models.FoodOrder{},
		&models.FoodOrderItem{},
	); err != nil {
		log.Printf("AutoMigrate warning: %v", err)
	}

	// Fix semua FK pada owner_uid agar selalu referensikan mitra_users(uid).
	// Ini mencakup: FK manual dari seeder dan FK auto-generated oleh GORM.
	config.DB.Exec(`ALTER TABLE mitra_stores DROP CONSTRAINT IF EXISTS mitra_stores_owner_uid_fkey`)
	config.DB.Exec(`ALTER TABLE mitra_stores DROP CONSTRAINT IF EXISTS fk_mitra_stores_owner`)
	config.DB.Exec(`ALTER TABLE mitra_stores DROP CONSTRAINT IF EXISTS mitra_stores_owner_uid_mitra_users_fkey`)
	if res := config.DB.Exec(`ALTER TABLE mitra_stores ADD CONSTRAINT fk_mitra_stores_owner FOREIGN KEY (owner_uid) REFERENCES mitra_users(uid) ON DELETE CASCADE`); res.Error != nil {
		log.Printf("FK fix warning: %v", res.Error)
	} else {
		log.Println("FK fk_mitra_stores_owner → mitra_users(uid) berhasil diperbaiki.")
	}

	userRepo := repositories.NewMitraUserRepository(config.DB)
	storeRepo := repositories.NewStoreRepository(config.DB)
	foodRepo := repositories.NewFoodRepository(config.DB)
	uploadRepo := repositories.NewUploadRepository(config.DB)
	otpRepo := repositories.NewOtpRepository(config.DB)
	cfgRepo := repositories.NewConfigRepository(config.DB)
	foodOrderRepo := repositories.NewFoodOrderRepository(config.DB)

	authHandler := handlers.NewAuthHandler(userRepo, otpRepo)
	storeHandler := handlers.NewStoreHandler(storeRepo)
	foodHandler := handlers.NewFoodHandler(foodRepo, storeRepo)
	uploadHandler := handlers.NewUploadHandler(uploadRepo, storeRepo, foodRepo)
	cfgHandler := handlers.NewConfigHandler(cfgRepo, userRepo)
	fcmSvc := fcm.New()
	foodOrderHandler := handlers.NewFoodOrderHandler(foodOrderRepo, foodRepo, storeRepo, userRepo, fcmSvc)

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.MaxMultipartMemory = 10 << 20 // 10 MB

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Backoffice-Secret"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	routes.SetupRoutes(r, authHandler, storeHandler, foodHandler, uploadHandler, cfgHandler, foodOrderHandler, userRepo)

	addr := ":" + config.AppConfig.ServerPort
	log.Printf("Talaku Mitra Food Service berjalan di %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
