-- Alter users.id column from UUID to TEXT to allow Firebase UIDs as primary keys
ALTER TABLE users ALTER COLUMN id TYPE TEXT USING id::text;

-- Add student-specific profile fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS student_id VARCHAR(20) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'student';
ALTER TABLE users ADD COLUMN IF NOT EXISTS major VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS minor VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS curriculum VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS pdpa_agreed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Create student_profiles table (extending user data for yellow card)
CREATE TABLE IF NOT EXISTS student_profiles (
    user_id    TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    advisor    VARCHAR(255),
    address    TEXT,
    phone      VARCHAR(20),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create yellow_card_subjects table
CREATE TABLE IF NOT EXISTS yellow_card_subjects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code        VARCHAR(20),
    name        VARCHAR(500),
    semester    VARCHAR(20),
    credits     VARCHAR(10),
    grade       VARCHAR(5),
    category    VARCHAR(100) NOT NULL,
    grp         VARCHAR(500),
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_yc_subjects_user ON yellow_card_subjects(user_id);

-- Create gpa_terms table for storing calculated/entered semester GPA summary
CREATE TABLE IF NOT EXISTS gpa_terms (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    semester VARCHAR(20) NOT NULL,
    ca       NUMERIC(5,2),
    cg       NUMERIC(5,2),
    gpa      NUMERIC(4,2),
    cax      NUMERIC(5,2),
    cgx      NUMERIC(5,2),
    gpax     NUMERIC(4,2),
    UNIQUE(user_id, semester)
);
