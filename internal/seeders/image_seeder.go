package seeders

import (
	"log"
	"talaku_mitra/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// cdnURL builds an Unsplash CDN URL from a photo ID
func cdnURL(photoID string, w, h int) string {
	return "https://images.unsplash.com/photo-" + photoID +
		"?w=" + itoa(w) + "&h=" + itoa(h) + "&fit=crop&q=80&auto=format"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// ─── Foto toko ───────────────────────────────────────────────────────────────

type storeImageData struct {
	storeName string
	logoID    string // Unsplash photo ID
	bannerID  string
}

var storeImages = []storeImageData{
	{
		storeName: "Warung Nasi Budi",
		logoID:    "OT-Wlz2Mn7w",                   // nasi goreng
		bannerID:  "1539755530862-00f623c00f52",     // indonesian food spread
	},
	{
		storeName: "Budi Juice Bar",
		logoID:    "JAJBmPXBxWE",                    // fresh juice
		bannerID:  "TgQkxQc-t_U",                   // juice bar counter
	},
	{
		storeName: "Dapur Siti",
		logoID:    "1680169590313-9a14f3cd8148",     // sundanese plating
		bannerID:  "PR3t-T_nTHQ",                   // asian food stall
	},
	{
		storeName: "Angkringan Andi",
		logoID:    "BfAkZvMrNSM",                   // street food stall
		bannerID:  "fuDES3VNEis",                   // night street food
	},
	{
		storeName: "Andi Bakso Spesial",
		logoID:    "90WRRmJhzsk",                   // meatball soup
		bannerID:  "1687425973269",                 // bakso bowl
	},
}

// ─── Foto makanan (maks 5 per item) ─────────────────────────────────────────

type foodImageData struct {
	foodName   string
	storeName  string
	photoIDs   []string // Unsplash photo IDs, urutan = sort_order
}

var foodImages = []foodImageData{
	// ── Warung Nasi Budi ────────────────────────────────────────────────────
	{
		foodName:  "Nasi Gudeg Komplit",
		storeName: "Warung Nasi Budi",
		photoIDs: []string{
			"o6Oq7rBMqVc",               // gudeg / nasi komplit
			"EdX2lJKAPWM",               // nasi dengan lauk
			"1613653739328-e86ebd77c9c8", // indonesian food plate
		},
	},
	{
		foodName:  "Nasi Ayam Bakar",
		storeName: "Warung Nasi Budi",
		photoIDs: []string{
			"g0dBbrGmMe0",               // grilled chicken rice
			"H1OC8oI5R5w",               // nasi ayam
			"1534939561126-855b8675edd7", // chicken dish
		},
	},
	{
		foodName:  "Nasi Tempe Orek",
		storeName: "Warung Nasi Budi",
		photoIDs: []string{
			"rQX9eVpSFz8",               // tempe dish
			"XciY4hwqnNk",               // rice with sides
		},
	},
	{
		foodName:  "Es Teh Manis",
		storeName: "Warung Nasi Budi",
		photoIDs: []string{
			"ccD0SOTmSwY",               // iced tea
			"_bQxQlLpoVY",               // cold drinks
		},
	},
	{
		foodName:  "Sayur Lodeh",
		storeName: "Warung Nasi Budi",
		photoIDs: []string{
			"1622572771591-6ca7813cc39d", // vegetable soup
			"1562607635-4608ff48a859",    // indonesian vegetables
		},
	},

	// ── Budi Juice Bar ──────────────────────────────────────────────────────
	{
		foodName:  "Jus Alpukat",
		storeName: "Budi Juice Bar",
		photoIDs: []string{
			"5aOzeDw_hcc",   // avocado juice
			"QD4yCjlD44A",   // green smoothie
			"ckilYix8R3U",   // creamy drink
		},
	},
	{
		foodName:  "Jus Mangga",
		storeName: "Budi Juice Bar",
		photoIDs: []string{
			"JWfcm1stQuo",   // mango juice yellow
			"zmeFA3kCqDs",   // tropical juice
		},
	},
	{
		foodName:  "Smoothie Bowl Stroberi",
		storeName: "Budi Juice Bar",
		photoIDs: []string{
			"_xRpRmF0Xl8",  // smoothie bowl
			"w2WBGMsORc0",  // berry bowl with granola
			"-P1KmzcJtN8",  // breakfast bowl
			"zc-rZTYKGzc",  // colorful bowl
		},
	},
	{
		foodName:  "Jus Wortel Jeruk",
		storeName: "Budi Juice Bar",
		photoIDs: []string{
			"0bWiLu_CExw",   // orange juice
			"xnbMrRWzSa8",   // carrot juice
		},
	},
	{
		foodName:  "Es Kelapa Muda",
		storeName: "Budi Juice Bar",
		photoIDs: []string{
			"bOoIlSsYN5g",   // coconut drink
			"87sqRxRPXZM",   // tropical coconut
			"PJBdtB2DSL8",   // fresh coconut water
		},
	},

	// ── Dapur Siti ──────────────────────────────────────────────────────────
	{
		foodName:  "Nasi Timbel Komplit",
		storeName: "Dapur Siti",
		photoIDs: []string{
			"1681378128359-a5c2492a3535", // banana leaf rice
			"1655740005902-2436216b82b8", // sundanese set
			"dSm9IDTYhr8",                // rice set meal
		},
	},
	{
		foodName:  "Karedok",
		storeName: "Dapur Siti",
		photoIDs: []string{
			"1682139710677-cb02f6bc4211", // vegetable salad peanut sauce
			"1666239308347-4292ea2ff777", // fresh salad asian
		},
	},
	{
		foodName:  "Pepes Ikan Mas",
		storeName: "Dapur Siti",
		photoIDs: []string{
			"1529563021893-cc83c992d75d", // fish banana leaf
			"1645696301019-35adcc18fc21", // grilled wrapped fish
		},
	},
	{
		foodName:  "Soto Bandung",
		storeName: "Dapur Siti",
		photoIDs: []string{
			"JXYdIWqA5IM",               // clear soup bowl
			"_jyB1ndDFQE",               // indonesian soup
			"1680674814945-7945d913319c", // soto bowl
		},
	},
	{
		foodName:  "Bajigur",
		storeName: "Dapur Siti",
		photoIDs: []string{
			"Lwicl8B_u4E",   // warm drink
			"f4PxhZIoqGQ",   // traditional hot drink
		},
	},

	// ── Angkringan Andi ─────────────────────────────────────────────────────
	{
		foodName:  "Nasi Kucing Teri",
		storeName: "Angkringan Andi",
		photoIDs: []string{
			"pbc2wXbQYpI",               // small rice wrap
			"1584455486010-760bd0b28fc2", // street food rice
		},
	},
	{
		foodName:  "Nasi Kucing Gudeg",
		storeName: "Angkringan Andi",
		photoIDs: []string{
			"i5XurHSjE1M",  // rice portion small
			"oT7_v-I0hHg",  // rice with topping
		},
	},
	{
		foodName:  "Sate Usus",
		storeName: "Angkringan Andi",
		photoIDs: []string{
			"GobbcGY38z4",  // satay skewer
			"J7S27AikvcE",  // grilled skewers
			"da-3hPifaeE",  // chicken satay
		},
	},
	{
		foodName:  "Sate Kulit Ayam",
		storeName: "Angkringan Andi",
		photoIDs: []string{
			"_zBGwx9VwuE",  // crispy chicken satay
			"f21Q33IcpyA",  // grilled chicken skin
			"0gcXl38ZLGA",  // satay peanut sauce
		},
	},
	{
		foodName:  "Wedang Jahe",
		storeName: "Angkringan Andi",
		photoIDs: []string{
			"jCgFnEjTRwI",  // ginger drink warm
			"BJTv0jylCeI",  // traditional warm drink
		},
	},
	{
		foodName:  "Tahu Bacem",
		storeName: "Angkringan Andi",
		photoIDs: []string{
			"bh9xEVaSxAE",  // fried tofu
			"eGXHsq0c36g",  // indonesian fried snack
		},
	},

	// ── Andi Bakso Spesial ───────────────────────────────────────────────────
	{
		foodName:  "Bakso Spesial",
		storeName: "Andi Bakso Spesial",
		photoIDs: []string{
			"1687425973283",  // bakso komplit
			"1747317368514",  // meatball soup big bowl
			"gHgwI8X_fl8",    // noodle meatball
			"UQL0dJGBFhM",    // soup bowl garnish
		},
	},
	{
		foodName:  "Bakso Biasa",
		storeName: "Andi Bakso Spesial",
		photoIDs: []string{
			"1696884422000",  // bakso bowl
			"1687426163461",  // meatball broth
			"y-wM_h_27Fg",    // simple noodle soup
		},
	},
	{
		foodName:  "Mie Ayam Bakso",
		storeName: "Andi Bakso Spesial",
		photoIDs: []string{
			"8Scd_34vdsw",   // chicken noodle
			"YEggQijngKc",   // noodle bowl asian
			"HTpiHBRoBIc",   // mie ayam
		},
	},
	{
		foodName:  "Bakso Goreng",
		storeName: "Andi Bakso Spesial",
		photoIDs: []string{
			"7R-S7x8gl4k",   // fried meatball
			"eMi4avHvImY",   // crispy fried snack
		},
	},
	{
		foodName:  "Es Jeruk",
		storeName: "Andi Bakso Spesial",
		photoIDs: []string{
			"HG0q7WcCKCQ",   // iced orange drink
			"c2f67maFA88",   // fresh citrus drink
			"--ZnV294AaQ",   // cold orange juice
		},
	},
}

// ─── Runner ──────────────────────────────────────────────────────────────────

func seedImages(db *gorm.DB) {
	log.Println("[Seeder] Memulai seeding images...")

	// 1. Seed logo & banner toko
	for _, si := range storeImages {
		var store models.Store
		if err := db.Where("name = ? AND deleted_at IS NULL", si.storeName).First(&store).Error; err != nil {
			log.Printf("[Seeder][Image] Store '%s' tidak ditemukan, skip.", si.storeName)
			continue
		}

		logoURL := cdnURL(si.logoID, 400, 400)
		bannerURL := cdnURL(si.bannerID, 1200, 400)

		updates := map[string]interface{}{
			"logo_url":   logoURL,
			"banner_url": bannerURL,
		}
		if err := db.Table("mitra_stores").Where("id = ?", store.ID).Updates(updates).Error; err != nil {
			log.Printf("[Seeder][Image] Gagal update store '%s': %v", si.storeName, err)
			continue
		}
		log.Printf("[Seeder][Image] Logo & banner toko '%s' di-set.", si.storeName)
	}

	// 2. Seed foto makanan
	count := 0
	for _, fi := range foodImages {
		var food models.Food
		err := db.
			Joins("JOIN mitra_stores ON mitra_foods.store_id = mitra_stores.id").
			Where("mitra_foods.name = ? AND mitra_stores.name = ? AND mitra_foods.deleted_at IS NULL", fi.foodName, fi.storeName).
			First(&food).Error
		if err != nil {
			log.Printf("[Seeder][Image] Food '%s' tidak ditemukan, skip.", fi.foodName)
			continue
		}

		// Cek apakah sudah ada foto
		var existing int64
		db.Model(&models.FoodImage{}).Where("food_id = ?", food.ID).Count(&existing)
		if existing > 0 {
			log.Printf("[Seeder][Image] Food '%s' sudah punya foto, skip.", fi.foodName)
			continue
		}

		// Insert foto (maks 5)
		limit := fi.photoIDs
		if len(limit) > models.MaxFoodImages {
			limit = limit[:models.MaxFoodImages]
		}

		for order, photoID := range limit {
			imgURL := cdnURL(photoID, 800, 600)
			img := models.FoodImage{
				ID:        uuid.New().String(),
				FoodID:    food.ID,
				URL:       imgURL,
				SortOrder: order,
			}
			if err := db.Create(&img).Error; err != nil {
				log.Printf("[Seeder][Image] Gagal insert foto '%s' order %d: %v", fi.foodName, order, err)
				continue
			}
			count++
		}

		// Set image_url utama ke foto pertama
		if len(limit) > 0 {
			mainURL := cdnURL(limit[0], 800, 600)
			db.Table("mitra_foods").Where("id = ?", food.ID).Update("image_url", mainURL)
		}

		log.Printf("[Seeder][Image] %d foto di-seed untuk '%s'.", len(limit), fi.foodName)
	}

	log.Printf("[Seeder][Image] Total %d foto makanan di-seed.", count)
}
