-- Migration: 0004_aksornsara_article_cms.down.sql
-- Description: Rollback AksornSara Article Management CMS features

-- ============================================================================
-- 1. DROP INDEXES
-- ============================================================================
DROP INDEX IF EXISTS idx_cms_user_roles_role;
DROP INDEX IF EXISTS idx_articles_author;
DROP INDEX IF EXISTS idx_articles_status_pubdate;

-- Revert idx_articles_slug to non-unique index if needed
DROP INDEX IF EXISTS idx_articles_slug;
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);

-- ============================================================================
-- 2. DROP TAXONOMY AND RELATIONAL TABLES
-- ============================================================================
DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS article_categories;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS categories;

-- ============================================================================
-- 3. DROP CMS ROLES & PERMISSION TABLES
-- ============================================================================
DROP TABLE IF EXISTS cms_user_roles;
DROP TABLE IF EXISTS cms_roles;

-- ============================================================================
-- 4. REVERT articles COLUMNS & STATUS ENUM
-- ============================================================================
ALTER TABLE articles DROP CONSTRAINT IF EXISTS fk_articles_published_by;
ALTER TABLE articles DROP COLUMN IF EXISTS published_by;
ALTER TABLE articles DROP COLUMN IF EXISTS writer_credit;
ALTER TABLE articles DROP COLUMN IF EXISTS proofreader_credit;
ALTER TABLE articles DROP COLUMN IF EXISTS illustrator_credit;
ALTER TABLE articles DROP COLUMN IF EXISTS meta_title;
ALTER TABLE articles DROP COLUMN IF EXISTS legacy_url;
ALTER TABLE articles DROP COLUMN IF EXISTS legacy_year;

-- Revert status enum back to VARCHAR(50)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'status' AND udt_name = 'article_status'
    ) THEN
        ALTER TABLE articles ALTER COLUMN status DROP DEFAULT;
        ALTER TABLE articles ALTER COLUMN status TYPE VARCHAR(50) USING status::text;
        ALTER TABLE articles ALTER COLUMN status SET DEFAULT 'draft';
    END IF;
END$$;

DROP TYPE IF EXISTS article_status;

