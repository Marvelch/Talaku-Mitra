package seeders

import (
	"log"
	"talaku_mitra/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func seedUsers(db *gorm.DB) []models.MitraUser {
	log.Println("[Seeder] Memulai seeding mitra_users...")

	type rawUser struct {
		fullName string
		email    string
		password string
		phone    string
	}

	data := []rawUser{
		{"Budi Santoso", "budi@talaku.id", "password123", "081234567890"},
		{"Siti Rahayu", "siti@talaku.id", "password123", "082345678901"},
		{"Andi Wijaya", "andi@talaku.id", "password123", "083456789012"},
	}

	var created []models.MitraUser

	for _, d := range data {
		var existing models.MitraUser
		if err := db.Where("email = ? AND deleted_at IS NULL", d.email).First(&existing).Error; err == nil {
			log.Printf("[Seeder] MitraUser %s sudah ada, skip.", d.email)
			created = append(created, existing)
			continue
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(d.password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("[Seeder] Gagal hash password untuk %s: %v", d.email, err)
		}

		phone := d.phone
		verified := true

		user := models.MitraUser{
			UID:             uuid.New().String(),
			FullName:        d.fullName,
			Email:           d.email,
			PasswordHash:    string(hashed),
			PhoneNumber:     &phone,
			IsActive:        true,
			IsVerifiedPhone: &verified,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("[Seeder] Gagal membuat mitra user %s: %v", d.email, err)
		}

		log.Printf("[Seeder] MitraUser dibuat: %s (%s)", d.fullName, d.email)
		created = append(created, user)
	}

	log.Printf("[Seeder] %d mitra users di-seed.\n", len(created))
	return created
}
