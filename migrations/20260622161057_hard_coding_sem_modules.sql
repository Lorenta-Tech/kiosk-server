-- +goose Up

-- +goose StatementBegin                                                
CREATE OR REPLACE FUNCTION create_semesters_for_branch()
RETURNS TRIGGER AS $$
DECLARE
    sem INT;
BEGIN
    FOR sem IN 1..8 LOOP
        INSERT INTO branch_semesters (id, branch_id, semester_number)
        VALUES (gen_random_uuid(), NEW.id, sem);
    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_create_semesters
    AFTER INSERT ON branches
    FOR EACH ROW
    EXECUTE FUNCTION create_semesters_for_branch();


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_modules_for_subject()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO modules (id, subject_id, module_number, title)
    VALUES
        (gen_random_uuid(), NEW.id, 1, 'Module 1'),
        (gen_random_uuid(), NEW.id, 2, 'Module 2'),
        (gen_random_uuid(), NEW.id, 3, 'Module 3'),
        (gen_random_uuid(), NEW.id, 4, 'Module 4'),
        (gen_random_uuid(), NEW.id, 5, 'Module 5'),
        (gen_random_uuid(), NEW.id, 6, 'Additional Resources');

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_create_modules
    AFTER INSERT ON subjects
    FOR EACH ROW
    EXECUTE FUNCTION create_modules_for_subject();


-- +goose Down

DROP TRIGGER IF EXISTS trg_create_semesters ON branches;
DROP FUNCTION IF EXISTS create_semesters_for_branch();

DROP TRIGGER IF EXISTS trg_create_modules ON subjects;
DROP FUNCTION IF EXISTS create_modules_for_subject();