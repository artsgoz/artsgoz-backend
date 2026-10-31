-- Migration: 0004_aksornsara_article_cms.up.sql
-- Description: Integrate AksornSara Article Management CMS features into Artsgoz

-- ============================================================================
-- 1. EXPAND EXISTING articles TABLE
-- ============================================================================

-- Ensure article_status enum exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'article_status') THEN
        CREATE TYPE article_status AS ENUM ('draft', 'published', 'archived');
    END IF;
END$$;

-- Add or convert status column to article_status enum
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'status'
    ) THEN
        ALTER TABLE articles ADD COLUMN status article_status NOT NULL DEFAULT 'draft';
    ELSE
        -- If status column exists but its data type is not the article_status enum
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'articles' AND column_name = 'status' AND udt_name <> 'article_status'
        ) THEN
            ALTER TABLE articles ALTER COLUMN status DROP DEFAULT;
            UPDATE articles 
            SET status = 'draft' 
            WHERE status NOT IN ('draft', 'published', 'archived') OR status IS NULL;
            ALTER TABLE articles ALTER COLUMN status TYPE article_status USING status::article_status;
            ALTER TABLE articles ALTER COLUMN status SET DEFAULT 'draft'::article_status;
            ALTER TABLE articles ALTER COLUMN status SET NOT NULL;
        END IF;
    END IF;
END$$;

-- Add CMS fields to articles table
ALTER TABLE articles ADD COLUMN IF NOT EXISTS excerpt TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS cover_image_url TEXT;

-- Add published_by with foreign key to users(id) ON DELETE SET NULL
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'published_by'
    ) THEN
        ALTER TABLE articles ADD COLUMN published_by UUID REFERENCES users(id) ON DELETE SET NULL;
    ELSE
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints 
            WHERE constraint_name = 'fk_articles_published_by' AND table_name = 'articles'
        ) THEN
            ALTER TABLE articles ADD CONSTRAINT fk_articles_published_by 
                FOREIGN KEY (published_by) REFERENCES users(id) ON DELETE SET NULL;
        END IF;
    END IF;
END$$;

ALTER TABLE articles ADD COLUMN IF NOT EXISTS writer_credit VARCHAR(255);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS proofreader_credit VARCHAR(255);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS illustrator_credit VARCHAR(255);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS meta_title VARCHAR(200);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS meta_description VARCHAR(300);
ALTER TABLE articles ADD COLUMN IF NOT EXISTS legacy_url TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS legacy_year VARCHAR(10);

-- ============================================================================
-- 2. CMS ROLES & PERMISSIONS
-- ============================================================================

CREATE TABLE IF NOT EXISTS cms_roles (
    id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS cms_user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES cms_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- ============================================================================
-- 3. TAXONOMY (CATEGORIES, TAGS & RELATION TABLES)
-- ============================================================================

-- If categories was previously created with UUID id (e.g. from prototype migration 0003),
-- recreate it with BIGSERIAL to align with AksornSara taxonomy specification
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'categories' AND column_name = 'id' AND data_type = 'uuid'
    ) THEN
        DROP TABLE categories CASCADE;
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(120) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure slug is VARCHAR(120) if table already existed with VARCHAR(100)
ALTER TABLE categories ALTER COLUMN slug TYPE VARCHAR(120);

CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(80) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_categories (
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (article_id, category_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_article_categories_primary 
    ON article_categories(article_id) 
    WHERE is_primary = TRUE;

CREATE TABLE IF NOT EXISTS article_tags (
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);

-- ============================================================================
-- 4. INDEXES FOR PERFORMANCE
-- ============================================================================

-- Fast lookup for published articles ordered by publication date
CREATE INDEX IF NOT EXISTS idx_articles_status_pubdate 
    ON articles(status, published_at DESC) 
    WHERE status = 'published';

-- Ensure slug has a UNIQUE index
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'articles' AND indexname = 'idx_articles_slug'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_index i
            JOIN pg_class c ON c.oid = i.indexrelid
            WHERE c.relname = 'idx_articles_slug' AND i.indisunique
        ) THEN
            DROP INDEX idx_articles_slug;
            CREATE UNIQUE INDEX idx_articles_slug ON articles(slug);
        END IF;
    ELSE
        CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
    END IF;
END$$;

-- Index on author/writer column (checking author_id, writerId, or author)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'author_id'
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_articles_author ON articles(author_id);
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'writerId'
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_articles_author ON articles("writerId");
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'author'
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_articles_author ON articles(author);
    END IF;
END$$;

-- Index for querying users by role
CREATE INDEX IF NOT EXISTS idx_cms_user_roles_role ON cms_user_roles(role_id, user_id);

