ALTER TABLE subjects ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS student_subjects (
    user_id     TEXT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id  UUID         NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    completed   BOOLEAN      NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_student_subjects_user ON student_subjects(user_id);
