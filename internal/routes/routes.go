package routes

import (
	"talaku_mitra/internal/handlers"
	"talaku_mitra/internal/middleware"
	"talaku_mitra/internal/models"

	"github.com/gin-gonic/gin"
)

// userFinderForRoute is the subset of MitraUserRepository used by the route setup.
type userFinderForRoute interface {
	FindByUID(uid string) (*models.MitraUser, error)
}

func SetupRoutes(
	r *gin.Engine,
	authHandler *handlers.AuthHandler,
	storeHandler *handlers.StoreHandler,
	foodHandler *handlers.FoodHandler,
	uploadHandler *handlers.UploadHandler,
	cfgHandler *handlers.ConfigHandler,
	userRepo userFinderForRoute,
) {
	// Sajikan file upload secara statis
	r.Static("/uploads", "./uploads")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "talaku-mitra-food"})
	})

	v1 := r.Group("/api/v1")

	// Auth (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/verify-forgot-otp", authHandler.VerifyForgotPasswordOtp)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.POST("/verify-phone", authHandler.VerifyPhone)
		auth.POST("/resend-otp", authHandler.ResendOtp)
	}

	// Auth (protected)
	authProtected := v1.Group("/auth")
	authProtected.Use(middleware.AuthRequired())
	{
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.GET("/profile", authHandler.GetProfile)
	}

	// Public: service status
	v1.GET("/public/services", cfgHandler.GetServiceStatus)

	// Public browse
	v1.GET("/foods", foodHandler.GetAllFoods)
	v1.GET("/foods/random-picks", foodHandler.GetRandomPicks)
	v1.GET("/foods/:id", foodHandler.GetFood)
	v1.GET("/foods/:id/images", uploadHandler.GetFoodImages)
	v1.GET("/stores", storeHandler.GetStores)
	v1.GET("/stores/:id", storeHandler.GetStore)
	v1.GET("/stores/:id/foods", foodHandler.GetStoreFoods)

	// Mitra food only (JWT wajib + user harus masih ada di DB)
	mitra := v1.Group("")
	mitra.Use(middleware.AuthRequired())
	mitra.Use(middleware.UserExistsRequired(userRepo))
	{
		// Store management
		mitra.POST("/stores", storeHandler.CreateStore)
		mitra.GET("/stores/my", storeHandler.GetMyStores)
		mitra.PUT("/stores/:id", storeHandler.UpdateStore)
		mitra.DELETE("/stores/:id", storeHandler.DeleteStore)

		// Store photos (multipart/form-data, field: "logo" / "banner")
		mitra.POST("/stores/:id/logo", uploadHandler.UploadStoreLogo)
		mitra.POST("/stores/:id/banner", uploadHandler.UploadStoreBanner)

		// Food management (:id = store_id, :food_id = food id)
		mitra.POST("/stores/:id/foods", foodHandler.CreateFood)
		mitra.PUT("/stores/:id/foods/:food_id", foodHandler.UpdateFood)
		mitra.DELETE("/stores/:id/foods/:food_id", foodHandler.DeleteFood)

		// Food photos (multipart/form-data, field: "image") — max 5 per makanan
		mitra.POST("/stores/:id/foods/:food_id/images", uploadHandler.UploadFoodImage)
		mitra.DELETE("/stores/:id/foods/:food_id/images/:image_id", uploadHandler.DeleteFoodImage)
	}
}
