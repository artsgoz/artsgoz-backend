CREATE TABLE IF NOT EXISTS professors (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name           VARCHAR(255) NOT NULL,
    department     VARCHAR(255) NOT NULL,
    location       VARCHAR(255),
    achievements   TEXT[]       NOT NULL DEFAULT '{}',
    qualifications TEXT[]       NOT NULL DEFAULT '{}',
    courses        TEXT[]       NOT NULL DEFAULT '{}',
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_professors_department ON professors(department);
CREATE INDEX idx_professors_status     ON professors(status);
CREATE INDEX idx_professors_active     ON professors(is_active);
