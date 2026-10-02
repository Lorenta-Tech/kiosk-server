-- +goose Up

-- 1. Reassign tokens that cannot satisfy the new rules: rows whose token is not exactly
--    6 digits, and every row after the first of a duplicate group (the oldest row of a
--    group keeps its token). The column was never unique, so historic sessions may share
--    a token and the unique index below would otherwise fail to apply.
-- +goose StatementBegin
DO
$$
    DECLARE
        stale_ids UUID[];
        stale_id  UUID;
        candidate TEXT;
    BEGIN
        SELECT COALESCE(ARRAY_AGG(id), '{}')
        INTO stale_ids
        FROM (SELECT id
              FROM upload_sessions
              WHERE token !~ '^[0-9]{6}$'

              UNION

              SELECT id
              FROM (SELECT id,
                           ROW_NUMBER() OVER (PARTITION BY token ORDER BY created_at ASC, id ASC) AS position
                    FROM upload_sessions
                    WHERE token ~ '^[0-9]{6}$') ranked
              WHERE position > 1) broken;

        FOREACH stale_id IN ARRAY stale_ids
            LOOP
                LOOP
                    candidate := LPAD((FLOOR(RANDOM() * 900000) + 100000)::BIGINT::TEXT, 6, '0');
                    EXIT WHEN NOT EXISTS (SELECT 1 FROM upload_sessions WHERE token = candidate);
                END LOOP;

                UPDATE upload_sessions
                SET token = candidate
                WHERE id = stale_id;
            END LOOP;
    END
$$;
-- +goose StatementEnd

-- 2. A token is now exactly 6 digits and identifies exactly one session.
ALTER TABLE upload_sessions
    ALTER COLUMN token TYPE VARCHAR(6);

ALTER TABLE upload_sessions
    ADD CONSTRAINT upload_sessions_token_format_chk CHECK (token ~ '^[0-9]{6}$'),
    ADD CONSTRAINT upload_sessions_token_key UNIQUE (token);

-- +goose Down

ALTER TABLE upload_sessions
    DROP CONSTRAINT IF EXISTS upload_sessions_token_key,
    DROP CONSTRAINT IF EXISTS upload_sessions_token_format_chk;

ALTER TABLE upload_sessions
    ALTER COLUMN token TYPE VARCHAR(100);