-- +goose Up
ALTER TABLE upload_files
    DROP COLUMN page_range;

ALTER TABLE upload_files
    ADD COLUMN page_range TEXT[] DEFAULT NULL;

-- +goose Down
ALTER TABLE upload_files
    DROP COLUMN page_range;

ALTER TABLE upload_files
    ADD COLUMN page_range VARCHAR(100);