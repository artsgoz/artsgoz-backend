-- Seed Script: cms_roles.sql
-- Description: Insert default CMS roles for Artsgoz AksornSara Article Management

INSERT INTO cms_roles (role_name, description)
VALUES 
    ('super_admin', 'Super Administrator with full access to CMS and system settings'),
    ('editor', 'Editor with permission to review, edit, and publish articles'),
    ('writer', 'Writer with permission to draft and submit articles')
ON CONFLICT (role_name) DO NOTHING;

