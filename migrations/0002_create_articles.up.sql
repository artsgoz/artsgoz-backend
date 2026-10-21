CREATE TABLE IF NOT EXISTS articles (
    id                   UUID PRIMARY KEY,
    slug                 VARCHAR(255) NOT NULL UNIQUE,
    title                VARCHAR(500) NOT NULL,
    subtitle             TEXT,
    excerpt              TEXT,
    content              TEXT NOT NULL,
    cover_image_url      TEXT,
    image_url            TEXT,
    author_id            VARCHAR(255),
    author_name          VARCHAR(255),
    category             VARCHAR(100),
    categories           TEXT NOT NULL DEFAULT '[]',
    category_ids         TEXT NOT NULL DEFAULT '[]',
    tags                 TEXT NOT NULL DEFAULT '[]',
    status               VARCHAR(50) NOT NULL DEFAULT 'draft',
    visibility           VARCHAR(50) NOT NULL DEFAULT 'public',
    published_at         TIMESTAMPTZ,
    meta_description     TEXT,
    seo_meta_description TEXT,
    views_count          BIGINT NOT NULL DEFAULT 0,
    references           TEXT NOT NULL DEFAULT '[]',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at);
CREATE INDEX IF NOT EXISTS idx_articles_deleted_at ON articles(deleted_at);
