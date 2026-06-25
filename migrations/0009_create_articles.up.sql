CREATE TABLE IF NOT EXISTS articles (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(500) NOT NULL,
    excerpt      TEXT         NOT NULL DEFAULT '',
    image_url    TEXT         NOT NULL DEFAULT '',
    author       VARCHAR(255) NOT NULL DEFAULT '',
    category     VARCHAR(100) NOT NULL DEFAULT '',
    content      TEXT[]       NOT NULL DEFAULT '{}',
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_category ON articles(category);
CREATE INDEX idx_articles_active   ON articles(is_active);
