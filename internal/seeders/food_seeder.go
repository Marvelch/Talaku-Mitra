package seeders

import (
	"log"
	"talaku_mitra/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedFoods(db *gorm.DB, stores []models.Store) {
	log.Println("[Seeder] Memulai seeding foods...")

	if len(stores) == 0 {
		log.Println("[Seeder] Tidak ada store, skip seeding foods.")
		return
	}

	type foodData struct {
		storeIndex  int
		name        string
		description string
		price       float64
		category    string
		isRecommend bool
		stock       int
		status      models.FoodStatus
	}

	ptr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	rawFoods := []foodData{
		// === Warung Nasi Budi (index 0) ===
		{0, "Nasi Gudeg Komplit", "Nasi putih dengan gudeg, ayam, telur, krecek, dan sambal goreng krecek.", 22000, "Makanan Berat", true, 50, models.FoodStatusAvailable},
		{0, "Nasi Ayam Bakar", "Nasi putih dengan ayam bakar bumbu kecap, lalapan, dan sambal.", 25000, "Makanan Berat", true, 30, models.FoodStatusAvailable},
		{0, "Nasi Tempe Orek", "Nasi putih dengan tempe orek manis pedas dan sayur tumis.", 15000, "Makanan Berat", false, 40, models.FoodStatusAvailable},
		{0, "Es Teh Manis", "Teh manis segar dengan es batu.", 5000, "Minuman", false, 100, models.FoodStatusAvailable},
		{0, "Sayur Lodeh", "Sayur lodeh labu siam, kacang panjang, dan tempe dalam santan gurih.", 10000, "Sayur", false, 20, models.FoodStatusAvailable},

		// === Budi Juice Bar (index 1) ===
		{1, "Jus Alpukat", "Jus alpukat segar dengan susu kental manis, creamy dan menyehatkan.", 18000, "Minuman", true, 50, models.FoodStatusAvailable},
		{1, "Jus Mangga", "Jus mangga harum manis pilihan, dingin menyegarkan.", 15000, "Minuman", false, 50, models.FoodStatusAvailable},
		{1, "Smoothie Bowl Stroberi", "Bowl smoothie stroberi dengan topping granola, pisang, dan chia seed.", 32000, "Minuman", true, 20, models.FoodStatusAvailable},
		{1, "Jus Wortel Jeruk", "Kombinasi wortel dan jeruk segar, kaya vitamin C.", 16000, "Minuman", false, 30, models.FoodStatusAvailable},
		{1, "Es Kelapa Muda", "Es kelapa muda asli dengan air dan daging kelapa yang segar.", 14000, "Minuman", false, 25, models.FoodStatusAvailable},

		// === Dapur Siti (index 2) ===
		{2, "Nasi Timbel Komplit", "Nasi timbel bungkus daun pisang dengan ayam goreng, tempe, tahu, dan lalapan.", 28000, "Makanan Berat", true, 30, models.FoodStatusAvailable},
		{2, "Karedok", "Salad sayuran segar dengan bumbu kacang khas Sunda.", 12000, "Sayur", true, 25, models.FoodStatusAvailable},
		{2, "Pepes Ikan Mas", "Ikan mas bumbu rempah dibungkus daun pisang dan dibakar.", 20000, "Lauk", false, 20, models.FoodStatusAvailable},
		{2, "Soto Bandung", "Soto daging sapi dengan lobak dan kacang kedelai, kuah bening.", 22000, "Sup", false, 15, models.FoodStatusAvailable},
		{2, "Bajigur", "Minuman tradisional Sunda dari gula aren dan santan, hangat dan nikmat.", 10000, "Minuman", false, 30, models.FoodStatusAvailable},

		// === Angkringan Andi (index 3) ===
		{3, "Nasi Kucing Teri", "Nasi putih porsi kecil dengan teri balado, cocok untuk ngemil malam.", 5000, "Makanan Ringan", true, 100, models.FoodStatusAvailable},
		{3, "Nasi Kucing Gudeg", "Nasi putih porsi kecil dengan gudeg manis khas Jogja.", 5000, "Makanan Ringan", false, 100, models.FoodStatusAvailable},
		{3, "Sate Usus", "Sate usus ayam bumbu kecap manis, dibakar sampai matang.", 3000, "Sate", false, 80, models.FoodStatusAvailable},
		{3, "Sate Kulit Ayam", "Sate kulit ayam renyah dengan bumbu kacang pedas.", 3000, "Sate", true, 80, models.FoodStatusAvailable},
		{3, "Wedang Jahe", "Minuman jahe hangat dengan gula jawa, menghangatkan badan.", 6000, "Minuman", false, 50, models.FoodStatusAvailable},
		{3, "Tahu Bacem", "Tahu bacem manis gurih, digoreng crispy.", 3000, "Gorengan", false, 60, models.FoodStatusAvailable},

		// === Andi Bakso Spesial (index 4) ===
		{4, "Bakso Spesial", "Bakso sapi ukuran besar dengan urat, tendon, dan sumsum dalam kuah bening.", 28000, "Makanan Berat", true, 40, models.FoodStatusAvailable},
		{4, "Bakso Biasa", "Bakso sapi standar dengan mie kuning dan bihun, kuah gurih.", 18000, "Makanan Berat", false, 50, models.FoodStatusAvailable},
		{4, "Mie Ayam Bakso", "Mie ayam segar dengan topping bakso sapi dan pangsit rebus.", 20000, "Makanan Berat", false, 35, models.FoodStatusAvailable},
		{4, "Bakso Goreng", "Bakso sapi digoreng crispy, cocok untuk camilan.", 15000, "Makanan Ringan", true, 30, models.FoodStatusAvailable},
		{4, "Es Jeruk", "Jeruk peras segar dengan es batu, menyegarkan setelah makan bakso.", 8000, "Minuman", false, 50, models.FoodStatusAvailable},
	}

	count := 0
	for _, f := range rawFoods {
		if f.storeIndex >= len(stores) {
			continue
		}
		storeID := stores[f.storeIndex].ID

		var existing models.Food
		if err := db.Where("name = ? AND store_id = ? AND deleted_at IS NULL", f.name, storeID).First(&existing).Error; err == nil {
			log.Printf("[Seeder] Food '%s' di store index %d sudah ada, skip.", f.name, f.storeIndex)
			continue
		}

		stock := f.stock
		food := models.Food{
			ID:          uuid.New().String(),
			StoreID:     storeID,
			Name:        f.name,
			Description: ptr(f.description),
			Price:       f.price,
			Category:    ptr(f.category),
			IsRecommend: f.isRecommend,
			Stock:       intPtr(stock),
			Status:      f.status,
		}

		if err := db.Create(&food).Error; err != nil {
			log.Fatalf("[Seeder] Gagal membuat food '%s': %v", f.name, err)
		}

		count++
		log.Printf("[Seeder] Food dibuat: %s - Rp%.0f (%s)", f.name, f.price, f.category)
	}

	log.Printf("[Seeder] %d foods berhasil di-seed.\n", count)
}
