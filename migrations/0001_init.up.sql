CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY,
    phone         TEXT        NOT NULL DEFAULT '',
    email         TEXT        NOT NULL DEFAULT '',
    password_hash TEXT        NOT NULL,
    type          TEXT        NOT NULL DEFAULT 'person',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_phone ON users (phone) WHERE phone <> '';
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email) WHERE email <> '';

CREATE TABLE IF NOT EXISTS user_roles
(
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role    TEXT   NOT NULL,
    PRIMARY KEY (user_id, role)
);

CREATE TABLE IF NOT EXISTS ads
(
    id                BIGSERIAL PRIMARY KEY,
    seller_id         BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    brand             TEXT        NOT NULL,
    model             TEXT        NOT NULL,
    year              INT         NOT NULL,
    mileage           INT         NOT NULL,
    price             INT         NOT NULL,
    vin               TEXT        NOT NULL DEFAULT '',
    city              TEXT        NOT NULL,
    status            TEXT        NOT NULL DEFAULT 'draft',
    inspection_status TEXT        NOT NULL DEFAULT 'none',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_ads_status ON ads (status);
CREATE INDEX IF NOT EXISTS ix_ads_city ON ads (city);
CREATE INDEX IF NOT EXISTS ix_ads_brand ON ads (brand);

CREATE TABLE IF NOT EXISTS inspections
(
    id           BIGSERIAL PRIMARY KEY,
    ad_id        BIGINT      NOT NULL REFERENCES ads (id) ON DELETE CASCADE,
    seller_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    inspector_id BIGINT      NULL REFERENCES users (id) ON DELETE SET NULL,
    status       TEXT        NOT NULL DEFAULT 'requested',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_inspections_inspector ON inspections (inspector_id);
CREATE INDEX IF NOT EXISTS ix_inspections_ad ON inspections (ad_id);

CREATE TABLE IF NOT EXISTS reports
(
    id            BIGSERIAL PRIMARY KEY,
    inspection_id BIGINT      NOT NULL REFERENCES inspections (id) ON DELETE CASCADE,
    total_score   INT         NOT NULL DEFAULT 0,
    label         TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
