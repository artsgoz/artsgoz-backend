-- Migration: 0002_dashboard_analytics_and_audit_logs.down.sql
-- Description: Rollback dashboard analytics and audit logs migration

-- Drop indexes and activity_logs table
DROP INDEX IF EXISTS idx_activity_logs_entity;
DROP INDEX IF EXISTS idx_activity_logs_user_id;
DROP INDEX IF EXISTS idx_activity_logs_created_at;
DROP TABLE IF EXISTS activity_logs;

-- Drop indexes and article_views table
DROP INDEX IF EXISTS idx_article_views_article_id;
DROP INDEX IF EXISTS idx_article_views_date;
DROP TABLE IF EXISTS article_views;

-- Remove counter cache column from articles
ALTER TABLE articles DROP COLUMN IF EXISTS view_count;

