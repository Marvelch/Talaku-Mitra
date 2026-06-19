-- =============================================================
-- Seed: Logo, Banner Toko & Foto Makanan
-- Sumber foto: Unsplash (images.unsplash.com)
-- Catatan: Jalankan SETELAH seed_users, seed_stores, seed_foods
-- =============================================================

-- ─────────────────────────────────────────────────────────────
-- 1. Logo & Banner Toko
-- ─────────────────────────────────────────────────────────────

UPDATE mitra_stores SET
    logo_url   = 'https://images.unsplash.com/photo-OT-Wlz2Mn7w?w=400&h=400&fit=crop&q=80&auto=format',
    banner_url = 'https://images.unsplash.com/photo-1539755530862-00f623c00f52?w=1200&h=400&fit=crop&q=80&auto=format'
WHERE name = 'Warung Nasi Budi' AND deleted_at IS NULL;

UPDATE mitra_stores SET
    logo_url   = 'https://images.unsplash.com/photo-JAJBmPXBxWE?w=400&h=400&fit=crop&q=80&auto=format',
    banner_url = 'https://images.unsplash.com/photo-TgQkxQc-t_U?w=1200&h=400&fit=crop&q=80&auto=format'
WHERE name = 'Budi Juice Bar' AND deleted_at IS NULL;

UPDATE mitra_stores SET
    logo_url   = 'https://images.unsplash.com/photo-1680169590313-9a14f3cd8148?w=400&h=400&fit=crop&q=80&auto=format',
    banner_url = 'https://images.unsplash.com/photo-PR3t-T_nTHQ?w=1200&h=400&fit=crop&q=80&auto=format'
WHERE name = 'Dapur Siti' AND deleted_at IS NULL;

UPDATE mitra_stores SET
    logo_url   = 'https://images.unsplash.com/photo-BfAkZvMrNSM?w=400&h=400&fit=crop&q=80&auto=format',
    banner_url = 'https://images.unsplash.com/photo-fuDES3VNEis?w=1200&h=400&fit=crop&q=80&auto=format'
WHERE name = 'Angkringan Andi' AND deleted_at IS NULL;

UPDATE mitra_stores SET
    logo_url   = 'https://images.unsplash.com/photo-90WRRmJhzsk?w=400&h=400&fit=crop&q=80&auto=format',
    banner_url = 'https://images.unsplash.com/photo-1687425973269?w=1200&h=400&fit=crop&q=80&auto=format'
WHERE name = 'Andi Bakso Spesial' AND deleted_at IS NULL;

-- ─────────────────────────────────────────────────────────────
-- 2. Foto Makanan (INSERT ke mitra_food_images)
--    Menggunakan subquery untuk lookup food_id by name+store
-- ─────────────────────────────────────────────────────────────

-- Helper function untuk lookup food_id
-- Format: get_food_id(food_name, store_name)
-- Dijalankan inline via subquery

-- ── Warung Nasi Budi ─────────────────────────────────────────

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(),
       f.id,
       v.url,
       v.sort_order,
       NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-o6Oq7rBMqVc?w=800&h=600&fit=crop&q=80&auto=format',               0),
    ('https://images.unsplash.com/photo-EdX2lJKAPWM?w=800&h=600&fit=crop&q=80&auto=format',               1),
    ('https://images.unsplash.com/photo-1613653739328-e86ebd77c9c8?w=800&h=600&fit=crop&q=80&auto=format', 2)
) AS v(url, sort_order)
WHERE f.name = 'Nasi Gudeg Komplit' AND s.name = 'Warung Nasi Budi'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-o6Oq7rBMqVc?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Nasi Gudeg Komplit'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Warung Nasi Budi' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-g0dBbrGmMe0?w=800&h=600&fit=crop&q=80&auto=format',               0),
    ('https://images.unsplash.com/photo-H1OC8oI5R5w?w=800&h=600&fit=crop&q=80&auto=format',               1),
    ('https://images.unsplash.com/photo-1534939561126-855b8675edd7?w=800&h=600&fit=crop&q=80&auto=format', 2)
) AS v(url, sort_order)
WHERE f.name = 'Nasi Ayam Bakar' AND s.name = 'Warung Nasi Budi'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-g0dBbrGmMe0?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Nasi Ayam Bakar'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Warung Nasi Budi' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-rQX9eVpSFz8?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-XciY4hwqnNk?w=800&h=600&fit=crop&q=80&auto=format', 1)
) AS v(url, sort_order)
WHERE f.name = 'Nasi Tempe Orek' AND s.name = 'Warung Nasi Budi'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-rQX9eVpSFz8?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Nasi Tempe Orek'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Warung Nasi Budi' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-ccD0SOTmSwY?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-_bQxQlLpoVY?w=800&h=600&fit=crop&q=80&auto=format', 1)
) AS v(url, sort_order)
WHERE f.name = 'Es Teh Manis' AND s.name = 'Warung Nasi Budi'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-ccD0SOTmSwY?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Es Teh Manis'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Warung Nasi Budi' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ── Budi Juice Bar ────────────────────────────────────────────

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-5aOzeDw_hcc?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-QD4yCjlD44A?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo-ckilYix8R3U?w=800&h=600&fit=crop&q=80&auto=format', 2)
) AS v(url, sort_order)
WHERE f.name = 'Jus Alpukat' AND s.name = 'Budi Juice Bar'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-5aOzeDw_hcc?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Jus Alpukat'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Budi Juice Bar' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-JWfcm1stQuo?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-zmeFA3kCqDs?w=800&h=600&fit=crop&q=80&auto=format', 1)
) AS v(url, sort_order)
WHERE f.name = 'Jus Mangga' AND s.name = 'Budi Juice Bar'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-JWfcm1stQuo?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Jus Mangga'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Budi Juice Bar' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-_xRpRmF0Xl8?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-w2WBGMsORc0?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo--P1KmzcJtN8?w=800&h=600&fit=crop&q=80&auto=format', 2),
    ('https://images.unsplash.com/photo-zc-rZTYKGzc?w=800&h=600&fit=crop&q=80&auto=format',  3)
) AS v(url, sort_order)
WHERE f.name = 'Smoothie Bowl Stroberi' AND s.name = 'Budi Juice Bar'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-_xRpRmF0Xl8?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Smoothie Bowl Stroberi'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Budi Juice Bar' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ── Andi Bakso Spesial ────────────────────────────────────────

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-1687425973283?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-1747317368514?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo-gHgwI8X_fl8?w=800&h=600&fit=crop&q=80&auto=format',  2),
    ('https://images.unsplash.com/photo-UQL0dJGBFhM?w=800&h=600&fit=crop&q=80&auto=format',  3)
) AS v(url, sort_order)
WHERE f.name = 'Bakso Spesial' AND s.name = 'Andi Bakso Spesial'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-1687425973283?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Bakso Spesial'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Andi Bakso Spesial' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-1696884422000?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-1687426163461?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo-y-wM_h_27Fg?w=800&h=600&fit=crop&q=80&auto=format',  2)
) AS v(url, sort_order)
WHERE f.name = 'Bakso Biasa' AND s.name = 'Andi Bakso Spesial'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-1696884422000?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Bakso Biasa'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Andi Bakso Spesial' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ──

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-8Scd_34vdsw?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-YEggQijngKc?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo-HTpiHBRoBIc?w=800&h=600&fit=crop&q=80&auto=format', 2)
) AS v(url, sort_order)
WHERE f.name = 'Mie Ayam Bakso' AND s.name = 'Andi Bakso Spesial'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-8Scd_34vdsw?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Mie Ayam Bakso'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Andi Bakso Spesial' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

-- ── Angkringan Andi ───────────────────────────────────────────

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-GobbcGY38z4?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-J7S27AikvcE?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo-da-3hPifaeE?w=800&h=600&fit=crop&q=80&auto=format',  2)
) AS v(url, sort_order)
WHERE f.name = 'Sate Usus' AND s.name = 'Angkringan Andi'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-GobbcGY38z4?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Sate Usus'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Angkringan Andi' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;

INSERT INTO mitra_food_images (id, food_id, url, sort_order, created_at)
SELECT gen_random_uuid(), f.id, v.url, v.sort_order, NOW()
FROM mitra_foods f
JOIN mitra_stores s ON f.store_id = s.id
CROSS JOIN (VALUES
    ('https://images.unsplash.com/photo-_zBGwx9VwuE?w=800&h=600&fit=crop&q=80&auto=format', 0),
    ('https://images.unsplash.com/photo-f21Q33IcpyA?w=800&h=600&fit=crop&q=80&auto=format', 1),
    ('https://images.unsplash.com/photo-0gcXl38ZLGA?w=800&h=600&fit=crop&q=80&auto=format', 2)
) AS v(url, sort_order)
WHERE f.name = 'Sate Kulit Ayam' AND s.name = 'Angkringan Andi'
  AND f.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM mitra_food_images WHERE food_id = f.id);

UPDATE mitra_foods SET image_url = 'https://images.unsplash.com/photo-_zBGwx9VwuE?w=800&h=600&fit=crop&q=80&auto=format'
WHERE name = 'Sate Kulit Ayam'
  AND store_id = (SELECT id FROM mitra_stores WHERE name = 'Angkringan Andi' AND deleted_at IS NULL LIMIT 1)
  AND deleted_at IS NULL;
