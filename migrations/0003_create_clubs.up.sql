CREATE TABLE IF NOT EXISTS clubs (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    category    VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    instagram   VARCHAR(255),
    image_url   TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clubs_category ON clubs(category);
CREATE INDEX idx_clubs_active   ON clubs(is_active);
