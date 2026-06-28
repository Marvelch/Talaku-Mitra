-- Food Order Tables for Talaku Mitra
-- Jalankan setelah init.sql

CREATE TABLE IF NOT EXISTS food_orders (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID          NOT NULL,        -- FK ke users.uid (main service)
    driver_id        UUID,                          -- FK ke drivers.uid (main service), nullable
    store_id         UUID          NOT NULL REFERENCES mitra_stores(id),
    status           VARCHAR(30)   NOT NULL DEFAULT 'waiting_driver'
                         CHECK (status IN (
                             'waiting_driver',      -- menunggu driver menerima
                             'waiting_restaurant',  -- driver sudah terima, menunggu restoran
                             'preparing',           -- restoran konfirmasi, sedang disiapkan
                             'ready',               -- makanan siap diambil
                             'on_delivery',         -- driver menuju customer
                             'delivered',           -- terkirim
                             'cancelled'            -- dibatalkan
                         )),
    subtotal         NUMERIC(12,2) NOT NULL DEFAULT 0,
    delivery_fee     NUMERIC(12,2) NOT NULL DEFAULT 0,
    service_fee      NUMERIC(12,2) NOT NULL DEFAULT 0,
    total            NUMERIC(12,2) NOT NULL DEFAULT 0,
    delivery_address TEXT          NOT NULL,
    delivery_lat     DOUBLE PRECISION,
    delivery_lng     DOUBLE PRECISION,
    note             TEXT,
    vehicle_type_id  INTEGER,
    vehicle_type_name VARCHAR(50),
    -- revenue distribution
    driver_amount    NUMERIC(12,2),
    talaku_gross     NUMERIC(12,2),
    tax_amount       NUMERIC(12,2),
    talaku_net       NUMERIC(12,2),
    -- timestamps
    accepted_at      TIMESTAMPTZ,
    confirmed_at     TIMESTAMPTZ,
    delivered_at     TIMESTAMPTZ,
    cancelled_at     TIMESTAMPTZ,
    cancel_reason    TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_food_orders_user_id   ON food_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_food_orders_driver_id ON food_orders(driver_id);
CREATE INDEX IF NOT EXISTS idx_food_orders_store_id  ON food_orders(store_id);
CREATE INDEX IF NOT EXISTS idx_food_orders_status    ON food_orders(status);
CREATE INDEX IF NOT EXISTS idx_food_orders_created_at ON food_orders(created_at DESC);

CREATE TABLE IF NOT EXISTS food_order_items (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID          NOT NULL REFERENCES food_orders(id) ON DELETE CASCADE,
    food_id     UUID          NOT NULL REFERENCES mitra_foods(id),
    food_name   VARCHAR(150)  NOT NULL,
    food_price  NUMERIC(12,2) NOT NULL,
    quantity    INTEGER       NOT NULL DEFAULT 1,
    subtotal    NUMERIC(12,2) NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_food_order_items_order_id ON food_order_items(order_id);
