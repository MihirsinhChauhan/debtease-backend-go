-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users ADD COLUMN google_id VARCHAR(255);

CREATE UNIQUE INDEX idx_users_google_id ON users (google_id)
    WHERE google_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_google_id;

ALTER TABLE users DROP COLUMN IF EXISTS google_id;

-- Restore NOT NULL: fail if OAuth-only users exist; fix data before downgrade.
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
-- +goose StatementEnd
