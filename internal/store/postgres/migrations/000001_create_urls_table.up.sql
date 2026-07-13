CREATE TABLE urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_code VARCHAR(64) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    clicks BIGINT NOT NULL DEFAULT 0,
    last_accessed TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_urls_is_active ON urls(is_active);
