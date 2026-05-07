-- +goose Up
-- +goose StatementBegin
-- Extend calculation_sessions for onboarding funnel tracking
ALTER TABLE calculation_sessions
    ADD COLUMN IF NOT EXISTS user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS entry_point TEXT NULL;

COMMENT ON COLUMN calculation_sessions.entry_point IS 'Funnel entry: debt_health, loan_test, landing';

CREATE INDEX IF NOT EXISTS idx_calculation_sessions_user_id
    ON calculation_sessions (user_id);

-- Expand debt_type enum for product scope (personal_loan, credit_card already exist)
ALTER TYPE debt_type ADD VALUE 'home_loan';
ALTER TYPE debt_type ADD VALUE 'car_loan';
ALTER TYPE debt_type ADD VALUE 'consumer_emi';
ALTER TYPE debt_type ADD VALUE 'bnpl';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove columns from calculation_sessions (cannot remove enum values in PostgreSQL)
ALTER TABLE calculation_sessions
    DROP COLUMN IF EXISTS entry_point,
    DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS idx_calculation_sessions_user_id;
-- +goose StatementEnd
