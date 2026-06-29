package seeders

import (
	"log"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	log.Println("[Seeder] Menjalankan migrasi tabel...")
	runMigrations(db)

	log.Println("[Seeder] Menjalankan semua seeder...")
	users := seedUsers(db)
	stores := seedStores(db, users)
	seedFoods(db, stores)
	seedImages(db)

	log.Println("[Seeder] Semua seeder selesai dijalankan.")
}

func runMigrations(db *gorm.DB) {
	sqls := []string{
		// ── Tabel akun mitra (terpisah dari tabel users utama) ─────────────
		`CREATE TABLE IF NOT EXISTS mitra_users (
			uid                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			full_name                 VARCHAR(255) NOT NULL,
			email                     VARCHAR(255) NOT NULL,
			password_hash             TEXT         NOT NULL,
			phone_number              VARCHAR(20),
			is_verified_phone         BOOLEAN      DEFAULT FALSE,
			refresh_token             TEXT,
			password_reset_code_hash  TEXT,
			password_reset_expires_at TIMESTAMPTZ,
			last_login                TIMESTAMPTZ,
			is_active                 BOOLEAN      NOT NULL DEFAULT TRUE,
			created_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			deleted_at                TIMESTAMPTZ
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_mitra_users_email      ON mitra_users(email) WHERE deleted_at IS NULL`,
		`CREATE INDEX        IF NOT EXISTS idx_mitra_users_phone      ON mitra_users(phone_number) WHERE deleted_at IS NULL`,
		`CREATE INDEX        IF NOT EXISTS idx_mitra_users_deleted_at ON mitra_users(deleted_at)`,

		// ── Tabel toko mitra ────────────────────────────────────────────────
		// FK owner_uid → mitra_users (bukan users)
		`CREATE TABLE IF NOT EXISTS mitra_stores (
			id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
			owner_uid   UUID         NOT NULL REFERENCES mitra_users(uid) ON DELETE CASCADE,
			name        VARCHAR(150) NOT NULL,
			description TEXT,
			address     TEXT         NOT NULL,
			phone       VARCHAR(20),
			logo_url    VARCHAR(255),
			banner_url  VARCHAR(255),
			status      VARCHAR(20)  NOT NULL DEFAULT 'active'
			                CHECK (status IN ('active','inactive','closed')),
			open_time   VARCHAR(10),
			close_time  VARCHAR(10),
			latitude    DOUBLE PRECISION,
			longitude   DOUBLE PRECISION,
			rating      NUMERIC(3,2) NOT NULL DEFAULT 0,
			rating_count INTEGER     NOT NULL DEFAULT 0,
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			deleted_at  TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_stores_owner_uid  ON mitra_stores(owner_uid)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_stores_status     ON mitra_stores(status)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_stores_deleted_at ON mitra_stores(deleted_at)`,

		// Migrasi FK yang sudah ada: ubah referensi owner_uid dari users → mitra_users
		// (aman dijalankan berkali-kali karena menggunakan IF EXISTS / IF NOT EXISTS)
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.table_constraints
				WHERE constraint_name = 'mitra_stores_owner_uid_fkey'
				  AND table_name = 'mitra_stores'
			) THEN
				ALTER TABLE mitra_stores DROP CONSTRAINT mitra_stores_owner_uid_fkey;
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints
				WHERE constraint_name = 'mitra_stores_owner_uid_mitra_users_fkey'
				  AND table_name = 'mitra_stores'
			) THEN
				ALTER TABLE mitra_stores
					ADD CONSTRAINT mitra_stores_owner_uid_mitra_users_fkey
					FOREIGN KEY (owner_uid) REFERENCES mitra_users(uid) ON DELETE CASCADE;
			END IF;
		END $$`,

		// ── Tabel makanan ───────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS mitra_foods (
			id           UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
			store_id     UUID          NOT NULL REFERENCES mitra_stores(id) ON DELETE CASCADE,
			name         VARCHAR(150)  NOT NULL,
			description  TEXT,
			price        NUMERIC(12,2) NOT NULL,
			category     VARCHAR(50),
			image_url    VARCHAR(255),
			status       VARCHAR(20)   NOT NULL DEFAULT 'available'
			                 CHECK (status IN ('available','unavailable')),
			is_recommend BOOLEAN       NOT NULL DEFAULT FALSE,
			stock        INTEGER,
			created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			deleted_at   TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_foods_store_id   ON mitra_foods(store_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_foods_category   ON mitra_foods(category)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_foods_status     ON mitra_foods(status)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_foods_deleted_at ON mitra_foods(deleted_at)`,

		// ── Tabel foto makanan (max 5 per item) ────────────────────────────
		`CREATE TABLE IF NOT EXISTS mitra_food_images (
			id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
			food_id    UUID         NOT NULL REFERENCES mitra_foods(id) ON DELETE CASCADE,
			url        VARCHAR(255) NOT NULL,
			sort_order INTEGER      NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_food_images_food_id ON mitra_food_images(food_id)`,

		// ── OTP verifikasi mitra ────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS mitra_otp_verifications (
			id         BIGSERIAL    PRIMARY KEY,
			type       VARCHAR(50)  NOT NULL,
			phone      VARCHAR(20)  NOT NULL,
			code       VARCHAR(10)  NOT NULL,
			is_used    BOOLEAN      NOT NULL DEFAULT FALSE,
			expires_at TIMESTAMPTZ  NOT NULL,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_mitra_otp_phone ON mitra_otp_verifications(phone)`,

		// ── Kolom persetujuan backoffice (idempotent) ───────────────────────
		`ALTER TABLE mitra_users ADD COLUMN IF NOT EXISTS is_approved_food BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE mitra_users ADD COLUMN IF NOT EXISTS is_approved_mart BOOLEAN NOT NULL DEFAULT FALSE`,

		// ── Tabel pesanan makanan ────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS food_orders (
			id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id          UUID           NOT NULL,
			driver_id        UUID,
			store_id         UUID           NOT NULL REFERENCES mitra_stores(id) ON DELETE RESTRICT,
			status           VARCHAR(30)    NOT NULL DEFAULT 'waiting_restaurant',
			subtotal         NUMERIC(12,2)  NOT NULL DEFAULT 0,
			delivery_fee     NUMERIC(12,2)  NOT NULL DEFAULT 0,
			service_fee      NUMERIC(12,2)  NOT NULL DEFAULT 0,
			total            NUMERIC(12,2)  NOT NULL DEFAULT 0,
			delivery_address TEXT           NOT NULL,
			delivery_lat     DOUBLE PRECISION,
			delivery_lng     DOUBLE PRECISION,
			note             TEXT,
			vehicle_type_id  INTEGER,
			vehicle_type_name VARCHAR(50),
			driver_amount    NUMERIC(12,2),
			talaku_gross     NUMERIC(12,2),
			tax_amount       NUMERIC(12,2),
			talaku_net       NUMERIC(12,2),
			accepted_at      TIMESTAMPTZ,
			confirmed_at     TIMESTAMPTZ,
			delivered_at     TIMESTAMPTZ,
			cancelled_at     TIMESTAMPTZ,
			cancel_reason    TEXT,
			created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
			updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_food_orders_user_id   ON food_orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_food_orders_driver_id  ON food_orders(driver_id)`,
		`CREATE INDEX IF NOT EXISTS idx_food_orders_store_id   ON food_orders(store_id)`,
		`CREATE INDEX IF NOT EXISTS idx_food_orders_status     ON food_orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_food_orders_created_at ON food_orders(created_at DESC)`,

		`CREATE TABLE IF NOT EXISTS food_order_items (
			id          UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id    UUID           NOT NULL REFERENCES food_orders(id) ON DELETE CASCADE,
			food_id     UUID           NOT NULL,
			food_name   VARCHAR(150)   NOT NULL,
			food_price  NUMERIC(12,2)  NOT NULL,
			quantity    INTEGER        NOT NULL DEFAULT 1,
			subtotal    NUMERIC(12,2)  NOT NULL,
			created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_food_order_items_order_id ON food_order_items(order_id)`,

		// ── Seed config layanan mitra jika belum ada ────────────────────────
		`INSERT INTO app_configs (parameter_key, parameter_value, description, is_active)
		 SELECT 'SERVICE_MART_ENABLED', 'true', 'Layanan mart Talaku Mitra', true
		 WHERE NOT EXISTS (SELECT 1 FROM app_configs WHERE parameter_key = 'SERVICE_MART_ENABLED')`,
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			log.Fatalf("[Migrasi] Gagal: %v\nSQL: %s", err, sql)
		}
	}

	log.Println("[Migrasi] Selesai.")
}
