-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_cards
    ADD COLUMN IF NOT EXISTS last4 TEXT,
    ADD COLUMN IF NOT EXISTS credit_limit NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS issuer TEXT;

CREATE TYPE statement_upload_status AS ENUM ('uploaded', 'parsing', 'parsed', 'failed');

CREATE TYPE statement_failure_code AS ENUM (
    'password_required',
    'password_invalid',
    'issuer_unknown',
    'parse_incomplete',
    'invalid_file',
    'internal_error'
);

CREATE TYPE statement_source AS ENUM ('MANUAL_PDF', 'GMAIL', 'SMS', 'API');

CREATE TABLE statement_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE CASCADE,
    statement_id UUID,
    storage_path TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    checksum TEXT NOT NULL,
    mime TEXT NOT NULL,
    status statement_upload_status NOT NULL DEFAULT 'uploaded',
    failure_code statement_failure_code,
    error_message TEXT,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE credit_card_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE CASCADE,
    statement_start DATE,
    statement_end DATE,
    statement_date DATE,
    due_date DATE,
    statement_amount NUMERIC(12,2),
    minimum_due NUMERIC(12,2),
    previous_balance NUMERIC(12,2),
    payments_received NUMERIC(12,2),
    new_balance NUMERIC(12,2),
    credit_limit NUMERIC(12,2),
    available_credit NUMERIC(12,2),
    source statement_source NOT NULL DEFAULT 'MANUAL_PDF',
    upload_id UUID NOT NULL REFERENCES statement_uploads(id) ON DELETE CASCADE,
    issuer_detected TEXT,
    parser_variant TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE statement_uploads
    ADD CONSTRAINT statement_uploads_statement_id_fkey
    FOREIGN KEY (statement_id) REFERENCES credit_card_statements(id) ON DELETE SET NULL;

CREATE INDEX idx_statement_uploads_debt_id ON statement_uploads(debt_id);
CREATE INDEX idx_credit_card_statements_debt_id ON credit_card_statements(debt_id);
CREATE INDEX idx_credit_card_statements_upload_id ON credit_card_statements(upload_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credit_card_statements;
DROP TABLE IF EXISTS statement_uploads;
DROP TYPE IF EXISTS statement_source;
DROP TYPE IF EXISTS statement_failure_code;
DROP TYPE IF EXISTS statement_upload_status;

ALTER TABLE credit_cards
    DROP COLUMN IF EXISTS issuer,
    DROP COLUMN IF EXISTS credit_limit,
    DROP COLUMN IF EXISTS last4;
-- +goose StatementEnd
