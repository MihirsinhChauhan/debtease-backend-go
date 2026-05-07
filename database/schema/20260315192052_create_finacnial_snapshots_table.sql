-- +goose Up
-- +goose StatementBegin
-- Financial snapshots: single source of truth for user financial state at a point in time.
-- Stores raw inputs only; health score and classification are recomputed from this data.
CREATE TABLE financial_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Link to identity (at least one of user_id or session_id is set)
    user_id    UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID NULL REFERENCES calculation_sessions(session_id) ON DELETE CASCADE,

    -- Snapshot classification
    snapshot_type   TEXT NOT NULL, -- onboarding | auto_generated
    snapshot_source TEXT NOT NULL, -- user_input | calculated

    -- Income
    monthly_income NUMERIC(14, 2) NULL,
    other_income   NUMERIC(14, 2) NULL,

    -- Expenses
    fixed_expenses    NUMERIC(14, 2) NULL,
    variable_expenses NUMERIC(14, 2) NULL,

    -- Assets
    savings_cash      NUMERIC(14, 2) NULL,
    investments_value NUMERIC(14, 2) NULL,

    -- Debt aggregates
    total_monthly_emi  NUMERIC(14, 2) NULL,
    total_debt_balance NUMERIC(14, 2) NULL,

    -- Household
    dependents_count INTEGER NULL,

    -- Metadata
    trigger_event    TEXT NULL,
    snapshot_version INTEGER NOT NULL DEFAULT 1,
    extra_data       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_financial_snapshots_identity
        CHECK (user_id IS NOT NULL OR session_id IS NOT NULL)
);

CREATE INDEX idx_financial_snapshots_user_id
    ON financial_snapshots (user_id);
CREATE INDEX idx_financial_snapshots_session_id
    ON financial_snapshots (session_id);
CREATE INDEX idx_financial_snapshots_created_at
    ON financial_snapshots (created_at);
CREATE INDEX idx_financial_snapshots_snapshot_type
    ON financial_snapshots (snapshot_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS financial_snapshots;
-- +goose StatementEnd
