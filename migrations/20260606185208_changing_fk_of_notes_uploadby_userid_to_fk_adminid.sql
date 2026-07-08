-- +goose Up

ALTER TABLE notes
    DROP CONSTRAINT IF EXISTS notes_uploaded_by_fkey;

ALTER TABLE notes
    ADD CONSTRAINT notes_uploaded_by_fkey
    FOREIGN KEY (uploaded_by)
    REFERENCES dept_admins(id)
    ON DELETE CASCADE;

-- +goose Down

ALTER TABLE notes
    DROP CONSTRAINT IF EXISTS notes_uploaded_by_fkey;

ALTER TABLE notes
    ADD CONSTRAINT notes_uploaded_by_fkey
    FOREIGN KEY (uploaded_by)
    REFERENCES users(id);
