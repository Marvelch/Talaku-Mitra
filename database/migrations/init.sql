-- Talaku Mitra Food Service - Migration
-- Jalankan sekali ke database 'talaku' yang sudah ada.

-- Tambah kolom is_food_mitra ke tabel users yang sudah ada
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_food_mitra BOOLEAN NOT NULL DEFAULT FALSE;

-- mitra_stores
CREATE TABLE IF NOT EXISTS mitra_stores (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_uid   UUID        NOT NULL REFERENCES users(uid) ON DELETE CASCADE,
    name        VARCHAR(150) NOT NULL,
    description TEXT,
    address     TEXT        NOT NULL,
    phone       VARCHAR(20),
    logo_url    VARCHAR(255),
    banner_url  VARCHAR(255),
    status      VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','inactive','closed')),
    open_time   VARCHAR(10),
    close_time  VARCHAR(10),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_mitra_stores_owner_uid  ON mitra_stores(owner_uid);
CREATE INDEX IF NOT EXISTS idx_mitra_stores_status     ON mitra_stores(status);
CREATE INDEX IF NOT EXISTS idx_mitra_stores_deleted_at ON mitra_stores(deleted_at);

-- mitra_foods
CREATE TABLE IF NOT EXISTS mitra_foods (
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
);

CREATE INDEX IF NOT EXISTS idx_mitra_foods_store_id   ON mitra_foods(store_id);
CREATE INDEX IF NOT EXISTS idx_mitra_foods_category   ON mitra_foods(category);
CREATE INDEX IF NOT EXISTS idx_mitra_foods_status     ON mitra_foods(status);
CREATE INDEX IF NOT EXISTS idx_mitra_foods_deleted_at ON mitra_foods(deleted_at);

-- Koordinat toko
ALTER TABLE mitra_stores ADD COLUMN IF NOT EXISTS latitude  DOUBLE PRECISION;
ALTER TABLE mitra_stores ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

-- Foto makanan (maksimal 5 per item)
CREATE TABLE IF NOT EXISTS mitra_food_images (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    food_id    UUID         NOT NULL REFERENCES mitra_foods(id) ON DELETE CASCADE,
    url        VARCHAR(255) NOT NULL,
    sort_order INTEGER      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mitra_food_images_food_id ON mitra_food_images(food_id);
