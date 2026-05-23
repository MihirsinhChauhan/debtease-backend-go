-- +goose Up
-- +goose StatementBegin
-- Append-only domain ledger (financial transitions). Application INSERT-only.
CREATE TABLE financial_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NULL,
    amount NUMERIC(14, 2) NULL,
    currency TEXT NOT NULL DEFAULT 'INR',
    occurred_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    schema_version SMALLINT NOT NULL DEFAULT 1
);

CREATE INDEX idx_financial_events_user_occurred ON financial_events (user_id, occurred_at DESC);

CREATE INDEX idx_financial_events_user_type ON financial_events (user_id, event_type);

CREATE INDEX idx_financial_events_entity ON financial_events (entity_type, entity_id)
WHERE entity_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS financial_events;
-- +goose StatementEnd
