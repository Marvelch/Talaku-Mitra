package main

import (
	"log"
	"talaku_mitra/internal/config"
	"talaku_mitra/internal/seeders"
)

func main() {
	config.Load()
	config.ConnectDB()

	log.Println("=== Talaku Mitra Food - Database Seeder ===")

	seeders.Run(config.DB)

	log.Println("=== Seeder selesai ===")
}
