-- +goose Up

CREATE TABLE branches (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL UNIQUE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE branch_semesters (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id       UUID NOT NULL REFERENCES branches(id),
    semester_number INT NOT NULL CHECK (semester_number BETWEEN 1 AND 8),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (branch_id, semester_number)
);

CREATE TABLE subjects (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_semester_id  UUID NOT NULL REFERENCES branch_semesters(id),
    name                TEXT NOT NULL,
    subject_code        TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (branch_semester_id, subject_code)
);

CREATE TABLE modules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id    UUID NOT NULL REFERENCES subjects(id),
    module_number INT NOT NULL CHECK (module_number BETWEEN 1 AND 6),
    title         TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subject_id, module_number)
);

CREATE TABLE notes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id         UUID NOT NULL REFERENCES modules(id),
    uploaded_by       UUID NOT NULL REFERENCES users(id),
    title             TEXT NOT NULL,
    description       TEXT NOT NULL DEFAULT '',
    file_key          TEXT NOT NULL,
    file_type         TEXT NOT NULL CHECK (file_type IN ('pdf','ppt','docx','image')),
    original_filename TEXT NOT NULL,
    file_size_bytes   INT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','active','deleted')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE department_admins (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    branch_id  UUID NOT NULL REFERENCES branches(id),
    role       TEXT NOT NULL DEFAULT 'dept_admin',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, branch_id)
);

CREATE INDEX idx_notes_module_id
    ON notes(module_id);

CREATE INDEX idx_notes_status
    ON notes(status);

CREATE INDEX idx_modules_subject_id
    ON modules(subject_id);

CREATE INDEX idx_subjects_branch_sem_id
    ON subjects(branch_semester_id);

CREATE INDEX idx_branch_semesters_branch
    ON branch_semesters(branch_id);


-- +goose Down

DROP INDEX IF EXISTS idx_branch_semesters_branch;
DROP INDEX IF EXISTS idx_subjects_branch_sem_id;
DROP INDEX IF EXISTS idx_modules_subject_id;
DROP INDEX IF EXISTS idx_notes_status;
DROP INDEX IF EXISTS idx_notes_module_id;

DROP TABLE IF EXISTS department_admins;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS modules;
DROP TABLE IF EXISTS subjects;
DROP TABLE IF EXISTS branch_semesters;
DROP TABLE IF EXISTS branches;