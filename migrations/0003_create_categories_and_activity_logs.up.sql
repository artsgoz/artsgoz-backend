CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color       VARCHAR(50),
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);

CREATE TABLE IF NOT EXISTS article_activity_logs (
    id          UUID PRIMARY KEY,
    article_id  UUID NOT NULL,
    user_id     VARCHAR(255),
    user_name   VARCHAR(255),
    user_role   VARCHAR(50),
    action      VARCHAR(50) NOT NULL,
    details     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_article_activity_logs_article_id ON article_activity_logs(article_id);
CREATE INDEX IF NOT EXISTS idx_article_activity_logs_created_at ON article_activity_logs(created_at);

