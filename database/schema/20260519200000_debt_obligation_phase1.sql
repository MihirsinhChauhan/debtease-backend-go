-- +goose Up
-- +goose StatementBegin
-- Phase 1: obligation model enums + parallel columns + backfill (TDD debt_obligation §5, §11.2).
CREATE TYPE debt_instrument_type AS ENUM (
    'home_loan',
    'car_loan',
    'gold_loan',
    'personal_loan',
    'education_loan',
    'medical_loan',
    'consumer_emi',
    'bnpl',
    'salary_advance',
    'pay_later',
    'credit_card',
    'line_of_credit'
);

CREATE TYPE interest_model AS ENUM (
    'fixed',
    'floating',
    'zero_cost_emi',
    'revolving'
);

ALTER TABLE debts
    ADD COLUMN instrument_type debt_instrument_type,
    ADD COLUMN interest_model interest_model,
    ADD COLUMN monthly_obligation NUMERIC(14, 2),
    ADD COLUMN minimum_obligation NUMERIC(14, 2),
    ADD COLUMN remaining_months INTEGER,
    ADD COLUMN due_day INTEGER,
    ADD COLUMN is_secured BOOLEAN,
    ADD COLUMN is_revolving BOOLEAN,
    ADD COLUMN is_short_term BOOLEAN,
    ADD COLUMN is_interest_bearing BOOLEAN,
    ADD COLUMN fragmentation_weight NUMERIC(4, 2),
    ADD COLUMN stress_priority SMALLINT,
    ADD COLUMN treatment_classification TEXT,
    ADD COLUMN engine_version TEXT,
    ADD COLUMN policy_version TEXT,
    ADD COLUMN extra_metadata JSONB;

ALTER TABLE financial_snapshots
    ADD COLUMN unsecured_monthly_obligation NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN revolving_monthly_obligation NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN short_term_obligations NUMERIC(14, 2) NOT NULL DEFAULT 0;

ALTER TABLE credit_cards
    ADD COLUMN minimum_due NUMERIC(14, 2);

-- Backfill from legacy `debts.type` and loan EMI fields.
UPDATE debts d
SET
    instrument_type = CASE d.type::text
        WHEN 'credit_card' THEN 'credit_card'::debt_instrument_type
        WHEN 'personal_loan' THEN 'personal_loan'::debt_instrument_type
        WHEN 'student_loan' THEN 'education_loan'::debt_instrument_type
        WHEN 'mortgage' THEN 'home_loan'::debt_instrument_type
        WHEN 'medical' THEN 'medical_loan'::debt_instrument_type
      -- Map legacy "other" away from a catch-all product label (see TDD §4.1).
        WHEN 'other' THEN 'personal_loan'::debt_instrument_type
        WHEN 'home_loan' THEN 'home_loan'::debt_instrument_type
        WHEN 'car_loan' THEN 'car_loan'::debt_instrument_type
        WHEN 'consumer_emi' THEN 'consumer_emi'::debt_instrument_type
        WHEN 'bnpl' THEN 'bnpl'::debt_instrument_type
        ELSE 'personal_loan'::debt_instrument_type
    END,
    interest_model = CASE
        WHEN d.type::text = 'credit_card' THEN 'revolving'::interest_model
        WHEN d.type::text = 'consumer_emi' AND COALESCE(d.interest_rate, 0) = 0 THEN 'zero_cost_emi'::interest_model
        WHEN d.type::text IN ('home_loan', 'mortgage') THEN 'floating'::interest_model
        ELSE 'fixed'::interest_model
    END,
    monthly_obligation = COALESCE(d.emi_amount, d.min_payment),
    minimum_obligation = d.min_payment,
    remaining_months = CASE
        WHEN d.tenure_months IS NOT NULL AND d.months_paid IS NOT NULL
            THEN GREATEST(d.tenure_months - d.months_paid, 0)
        WHEN d.tenure_months IS NOT NULL THEN GREATEST(d.tenure_months, 0)
        ELSE NULL
    END,
    due_day = COALESCE(d.payment_due_day, EXTRACT(DAY FROM d.due_date)::integer),
    is_secured = CASE
        WHEN d.type::text IN ('home_loan', 'mortgage', 'car_loan') THEN TRUE
        ELSE FALSE
    END,
    is_revolving = (d.type::text = 'credit_card'),
    is_short_term = (d.type::text = 'bnpl'),
    is_interest_bearing = CASE
        WHEN d.type::text = 'bnpl' THEN FALSE
        WHEN d.type::text = 'consumer_emi' AND COALESCE(d.interest_rate, 0) = 0 THEN FALSE
        ELSE COALESCE(d.interest_rate, 0) > 0
    END,
    fragmentation_weight = CASE d.type::text
        WHEN 'bnpl' THEN 0.50
        WHEN 'credit_card' THEN 0.70
        WHEN 'consumer_emi' THEN 0.60
        WHEN 'personal_loan' THEN 0.85
        WHEN 'student_loan' THEN 0.80
        WHEN 'medical' THEN 0.85
        WHEN 'other' THEN 0.85
        WHEN 'car_loan' THEN 0.95
        WHEN 'home_loan' THEN 1.00
        WHEN 'mortgage' THEN 1.00
        ELSE 0.85
    END::numeric(4, 2),
    stress_priority = CASE d.type::text
        WHEN 'bnpl' THEN 1
        WHEN 'credit_card' THEN 2
        WHEN 'consumer_emi' THEN 2
        WHEN 'personal_loan' THEN 3
        WHEN 'student_loan' THEN 3
        WHEN 'medical' THEN 3
        WHEN 'other' THEN 3
        WHEN 'car_loan' THEN 4
        WHEN 'home_loan' THEN 4
        WHEN 'mortgage' THEN 4
        ELSE 3
    END::smallint,
    engine_version = '1',
    policy_version = '1';

-- Credit card: optional minimum_due from legacy min_payment on debt row.
UPDATE credit_cards cc
SET minimum_due = d.min_payment
FROM debts d
WHERE cc.debt_id = d.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_cards DROP COLUMN IF EXISTS minimum_due;

ALTER TABLE financial_snapshots
    DROP COLUMN IF EXISTS unsecured_monthly_obligation,
    DROP COLUMN IF EXISTS revolving_monthly_obligation,
    DROP COLUMN IF EXISTS short_term_obligations;

ALTER TABLE debts
    DROP COLUMN IF EXISTS instrument_type,
    DROP COLUMN IF EXISTS interest_model,
    DROP COLUMN IF EXISTS monthly_obligation,
    DROP COLUMN IF EXISTS minimum_obligation,
    DROP COLUMN IF EXISTS remaining_months,
    DROP COLUMN IF EXISTS due_day,
    DROP COLUMN IF EXISTS is_secured,
    DROP COLUMN IF EXISTS is_revolving,
    DROP COLUMN IF EXISTS is_short_term,
    DROP COLUMN IF EXISTS is_interest_bearing,
    DROP COLUMN IF EXISTS fragmentation_weight,
    DROP COLUMN IF EXISTS stress_priority,
    DROP COLUMN IF EXISTS treatment_classification,
    DROP COLUMN IF EXISTS engine_version,
    DROP COLUMN IF EXISTS policy_version,
    DROP COLUMN IF EXISTS extra_metadata;

DROP TYPE IF EXISTS interest_model;
DROP TYPE IF EXISTS debt_instrument_type;
-- +goose StatementEnd
