CREATE TABLE IF NOT EXISTS curricula (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name           VARCHAR(500) NOT NULL,
    year           INTEGER      NOT NULL,
    total_credits  INTEGER      NOT NULL,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS curriculum_categories (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    curriculum_id    UUID         NOT NULL REFERENCES curricula(id) ON DELETE CASCADE,
    category         VARCHAR(100) NOT NULL,
    required_credits INTEGER      NOT NULL,
    groups           TEXT[]       NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(curriculum_id, category)
);

CREATE TABLE IF NOT EXISTS subjects (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    curriculum_id UUID          NOT NULL REFERENCES curricula(id) ON DELETE CASCADE,
    code          VARCHAR(20)   NOT NULL,
    name_th       VARCHAR(500)  NOT NULL,
    name_en       VARCHAR(500)  NOT NULL,
    credits       NUMERIC(4,2)  NOT NULL,
    category      VARCHAR(100)  NOT NULL,
    semester      INTEGER,      -- 1-8
    grp           VARCHAR(500),  -- e.g. group name
    is_custom     BOOLEAN       NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subjects_curriculum ON subjects(curriculum_id);
CREATE INDEX idx_curriculum_categories_curr ON curriculum_categories(curriculum_id);
