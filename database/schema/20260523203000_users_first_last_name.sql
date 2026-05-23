-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN first_name VARCHAR(255);
ALTER TABLE users ADD COLUMN last_name VARCHAR(255);

UPDATE users SET
  first_name = CASE
    WHEN strpos(trim(name), ' ') > 0 THEN substring(trim(name) FROM 1 FOR strpos(trim(name), ' ') - 1)
    ELSE trim(name)
  END,
  last_name = CASE
    WHEN strpos(trim(name), ' ') > 0 THEN trim(substring(trim(name) FROM strpos(trim(name), ' ') + 1))
    ELSE ''
  END;

ALTER TABLE users ALTER COLUMN first_name SET NOT NULL;
ALTER TABLE users ALTER COLUMN last_name SET NOT NULL;
ALTER TABLE users ALTER COLUMN last_name SET DEFAULT '';

ALTER TABLE users DROP COLUMN name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN name VARCHAR(255);

UPDATE users SET
  name = CASE
    WHEN COALESCE(trim(last_name), '') = '' THEN trim(first_name)
    ELSE trim(first_name) || ' ' || trim(last_name)
  END;

ALTER TABLE users ALTER COLUMN name SET NOT NULL;

ALTER TABLE users DROP COLUMN last_name;
ALTER TABLE users DROP COLUMN first_name;
-- +goose StatementEnd
