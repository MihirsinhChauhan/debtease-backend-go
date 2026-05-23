-- +goose Up
-- +goose StatementBegin
-- Phase 4: drop loan-centric / risk columns; tighten obligation model (TDD §15).
UPDATE debts
SET monthly_obligation = COALESCE(monthly_obligation, min_payment)
WHERE monthly_obligation IS NULL;

UPDATE debts
SET minimum_obligation = COALESCE(minimum_obligation, min_payment)
WHERE minimum_obligation IS NULL;

UPDATE debts
SET due_day = COALESCE(due_day, payment_due_day, EXTRACT(DAY FROM due_date)::integer)
WHERE due_day IS NULL;

UPDATE debts
SET engine_version = COALESCE(engine_version, '1')
WHERE engine_version IS NULL;

UPDATE debts
SET policy_version = COALESCE(policy_version, '1')
WHERE policy_version IS NULL;

UPDATE debts
SET is_secured = COALESCE(is_secured, FALSE)
WHERE is_secured IS NULL;

UPDATE debts
SET is_revolving = COALESCE(is_revolving, FALSE)
WHERE is_revolving IS NULL;

UPDATE debts
SET is_short_term = COALESCE(is_short_term, FALSE)
WHERE is_short_term IS NULL;

UPDATE debts
SET is_interest_bearing = COALESCE(is_interest_bearing, TRUE)
WHERE is_interest_bearing IS NULL;

UPDATE debts
SET fragmentation_weight = COALESCE(fragmentation_weight, 0.85::numeric(4, 2))
WHERE fragmentation_weight IS NULL;

UPDATE debts
SET stress_priority = COALESCE(stress_priority, 3::smallint)
WHERE stress_priority IS NULL;

UPDATE debts
SET instrument_type = 'personal_loan'::debt_instrument_type
WHERE instrument_type IS NULL;

UPDATE debts
SET interest_model = 'fixed'::interest_model
WHERE interest_model IS NULL;

ALTER TABLE debts
    ALTER COLUMN monthly_obligation SET NOT NULL,
    ALTER COLUMN minimum_obligation SET NOT NULL,
    ALTER COLUMN is_secured SET NOT NULL,
    ALTER COLUMN is_revolving SET NOT NULL,
    ALTER COLUMN is_short_term SET NOT NULL,
    ALTER COLUMN is_interest_bearing SET NOT NULL,
    ALTER COLUMN fragmentation_weight SET NOT NULL,
    ALTER COLUMN stress_priority SET NOT NULL,
    ALTER COLUMN engine_version SET NOT NULL,
    ALTER COLUMN policy_version SET NOT NULL,
    ALTER COLUMN instrument_type SET NOT NULL,
    ALTER COLUMN interest_model SET NOT NULL;

ALTER TABLE debts
    DROP COLUMN IF EXISTS emi_amount,
    DROP COLUMN IF EXISTS tenure_months,
    DROP COLUMN IF EXISTS months_paid,
    DROP COLUMN IF EXISTS risk_level;

DROP TYPE IF EXISTS risk_level;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high');

ALTER TABLE debts
    ADD COLUMN risk_level risk_level,
    ADD COLUMN emi_amount DECIMAL(12, 2),
    ADD COLUMN tenure_months INTEGER,
    ADD COLUMN months_paid INTEGER DEFAULT 0;

UPDATE debts d
SET
    emi_amount = d.monthly_obligation,
    months_paid = 0;

ALTER TABLE debts ALTER COLUMN months_paid SET DEFAULT 0;
-- +goose StatementEnd
