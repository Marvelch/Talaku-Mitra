package seeders

import (
	"log"
	"talaku_mitra/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedStores(db *gorm.DB, users []models.MitraUser) []models.Store {
	log.Println("[Seeder] Memulai seeding stores...")

	if len(users) == 0 {
		log.Println("[Seeder] Tidak ada mitra user, skip.")
		return nil
	}

	type storeData struct {
		ownerIdx    int
		name        string
		description string
		address     string
		phone       string
		openTime    string
		closeTime   string
		lat         float64
		lng         float64
		rating      float64
		ratingCount int
	}

	data := []storeData{
		// ── Makassar, Pulau Sulawesi (~lat -5.14, lng 119.43) ─────────────────
		{0, "Warung Nasi Budi", "Warung nasi rumahan dengan menu masakan khas dengan cita rasa lezat dan terjangkau.", "Jl. Penghibur No. 10, Makassar", "081234567890", "07:00", "21:00", -5.1364, 119.4221, 4.8, 152},
		{0, "Budi Juice Bar", "Minuman segar jus buah dan smoothie bowl pilihan.", "Jl. Somba Opu No. 3, Makassar", "081234567891", "08:00", "20:00", -5.1420, 119.4180, 4.3, 64},
		{1, "Dapur Siti", "Spesialis masakan dengan cita rasa otentik dan bahan-bahan segar.", "Jl. Cendrawasih No. 25, Makassar", "082345678901", "09:00", "22:00", -5.1558, 119.4312, 4.6, 98},
		{2, "Angkringan Andi", "Angkringan khas dengan nasi kucing dan berbagai gorengan murah meriah.", "Jl. G. Bawakaraeng No. 5, Makassar", "083456789012", "16:00", "23:59", -5.1490, 119.4390, 4.1, 47},
		{2, "Andi Bakso Spesial", "Bakso sapi asli dengan kuah bening gurih, tersedia berbagai topping pilihan.", "Jl. Perintis Kemerdekaan No. 88, Makassar", "083456789013", "10:00", "20:00", -5.1250, 119.4850, 4.7, 210},

		// ── Pulau Salibabu, Kepulauan Talaud (~lat 4.25, lng 126.70) ──────────
		{0, "Warung Kita Salibabu", "Masakan rumahan khas Talaud, nasi campur dan ikan segar.", "Desa Salibabu, Kec. Salibabu, Talaud", "085211110001", "07:00", "20:00", 4.2520, 126.6980, 4.5, 38},
		{1, "Dapur Mama Ros", "Makanan khas Talaud: ikan kuah kuning, nasi jagung, dan gohu ikan.", "Desa Kalabat, Pulau Salibabu, Talaud", "085211110002", "08:00", "21:00", 4.2480, 126.7050, 4.9, 61},
		{2, "Kantin Pak Yusuf", "Kantin sederhana dengan menu harian berganti setiap hari.", "Pelabuhan Salibabu, Talaud", "085211110003", "06:00", "18:00", 4.2440, 126.6920, 4.0, 19},
	}

	var created []models.Store

	for _, s := range data {
		if s.ownerIdx >= len(users) {
			continue
		}
		ownerUID := users[s.ownerIdx].UID

		var existing models.Store
		if err := db.Where("name = ? AND owner_uid = ? AND deleted_at IS NULL", s.name, ownerUID).First(&existing).Error; err == nil {
			log.Printf("[Seeder] Store '%s' sudah ada, skip.", s.name)
			updates := map[string]interface{}{}
			if existing.Latitude == nil {
				updates["latitude"] = s.lat
				updates["longitude"] = s.lng
			}
			if existing.Rating == 0 {
				updates["rating"] = s.rating
				updates["rating_count"] = s.ratingCount
			}
			if len(updates) > 0 {
				db.Model(&existing).Updates(updates)
			}
			created = append(created, existing)
			continue
		}

		desc, phone, open, close, lat, lng := s.description, s.phone, s.openTime, s.closeTime, s.lat, s.lng
		store := models.Store{
			ID:          uuid.New().String(),
			OwnerUID:    ownerUID,
			Name:        s.name,
			Description: &desc,
			Address:     s.address,
			Phone:       &phone,
			OpenTime:    &open,
			CloseTime:   &close,
			Latitude:    &lat,
			Longitude:   &lng,
			Rating:      s.rating,
			RatingCount: s.ratingCount,
			Status:      models.StoreStatusActive,
		}

		if err := db.Create(&store).Error; err != nil {
			log.Fatalf("[Seeder] Gagal membuat store '%s': %v", s.name, err)
		}

		log.Printf("[Seeder] Store dibuat: %s (owner: %s)", s.name, users[s.ownerIdx].FullName)
		created = append(created, store)
	}

	log.Printf("[Seeder] %d stores di-seed.\n", len(created))
	return created
}
