-- Migration: 0002_dashboard_analytics_and_audit_logs.up.sql
-- Description: Add counter cache to articles, create article_views for time-series analytics, and activity_logs for admin audit trail

-- ============================================================================
-- 1. COUNTER CACHE IN ARTICLES TABLE
-- ============================================================================
ALTER TABLE articles ADD COLUMN IF NOT EXISTS view_count BIGINT NOT NULL DEFAULT 0;

-- Sync legacy views_count data if the column exists and view_count is still 0
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'articles' AND column_name = 'views_count'
    ) THEN
        UPDATE articles SET view_count = views_count WHERE view_count = 0 AND views_count > 0;
    END IF;
END$$;

-- ============================================================================
-- 2. ARTICLE VIEWS LOG TABLE (TIME-SERIES & PRIVACY-COMPLIANT)
-- ============================================================================
-- Note: articles.id is UUID in Artsgoz system, so article_id uses UUID
CREATE TABLE IF NOT EXISTS article_views (
    id BIGSERIAL PRIMARY KEY,
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    ip_hash VARCHAR(64),
    user_agent TEXT,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 3. AUDIT / ACTIVITY LOGS TABLE (ADMIN DASHBOARD)
-- ============================================================================
-- Clean up prototype table if previously created with UUID primary key
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'activity_logs' AND column_name = 'id' AND data_type = 'uuid'
    ) THEN
        DROP TABLE activity_logs CASCADE;
    END IF;
END$$;

-- Note: users.id is UUID in Artsgoz system, so user_id uses UUID
CREATE TABLE IF NOT EXISTS activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 4. INDEXES FOR DASHBOARD QUERY PERFORMANCE
-- ============================================================================
-- Time-series range queries by date and article (Dashboard trends)
CREATE INDEX IF NOT EXISTS idx_article_views_date ON article_views(viewed_at, article_id);

-- Foreign key lookup and CASCADE optimization for article_views
CREATE INDEX IF NOT EXISTS idx_article_views_article_id ON article_views(article_id);

-- Fast retrieval of latest activity logs for admin dashboard
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);

-- Foreign key lookup and SET NULL optimization for activity_logs
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);

-- Composite index for filtering activities by target entity
CREATE INDEX IF NOT EXISTS idx_activity_logs_entity ON activity_logs(entity_type, entity_id);

