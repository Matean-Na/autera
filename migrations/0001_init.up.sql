CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY,
    phone         TEXT        NOT NULL DEFAULT '',
    email         TEXT        NOT NULL DEFAULT '',
    password_hash TEXT        NOT NULL,
    type          TEXT        NOT NULL DEFAULT 'person',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    token_version BIGINT      NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_phone ON users (phone) WHERE phone <> '';
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email) WHERE email <> '';

CREATE TABLE IF NOT EXISTS user_roles
(
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role    TEXT   NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- refresh tokens (ротация + мультисессии по device_id)
CREATE TABLE user_refresh_tokens
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    jti        TEXT        NOT NULL UNIQUE,
    token_hash TEXT        NOT NULL,
    device_id  TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_refresh_tokens_user_device
    ON user_refresh_tokens(user_id, device_id);

CREATE INDEX idx_user_refresh_tokens_expires
    ON user_refresh_tokens(expires_at);

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
