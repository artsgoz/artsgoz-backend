CREATE TABLE IF NOT EXISTS documents (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(500) NOT NULL,
    details     TEXT         NOT NULL DEFAULT '',
    category    VARCHAR(100) NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'none',
    file_url    TEXT         NOT NULL,
    file_name   VARCHAR(255),
    file_size   BIGINT,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_category ON documents(category);
CREATE INDEX idx_documents_active   ON documents(is_active);
