-- +goose Up

DROP TABLE IF EXISTS department_admins;

CREATE TABLE dept_admins (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id     UUID NOT NULL REFERENCES branches(id),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'dept_admin',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS dept_admins;

CREATE TABLE department_admins (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    branch_id  UUID NOT NULL REFERENCES branches(id),
    role       TEXT NOT NULL DEFAULT 'dept_admin',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, branch_id)
);
