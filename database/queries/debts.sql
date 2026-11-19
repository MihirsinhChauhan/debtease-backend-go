-- name: CreateDebt :one
INSERT INTO debts (
  user_id, name, type, lender, principal_decimal,
  outstanding_balance, interest_rate, min_payment,
  due_date, payment_frequency, debt_taken,
  priority, notes,
  payment_due_day, moratorium_until, in_moratorium,
  instrument_type, interest_model,
  monthly_obligation, minimum_obligation, remaining_months, due_day,
  is_secured, is_revolving, is_short_term, is_interest_bearing,
  fragmentation_weight, stress_priority,
  treatment_classification, engine_version, policy_version, extra_metadata
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
  $12, $13, $14, $15, $16,
  $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32
)
RETURNING *;

-- name: ListUserDebts :many
SELECT *
FROM debts
WHERE user_id = $1
ORDER BY priority DESC, due_date ASC;

-- name: GetDebtByID :one
SELECT *
FROM debts
WHERE id = $1 AND user_id = $2;

-- name: UpdateDebtDetails :one
UPDATE debts
SET name = $3,
    lender = $4,
    principal_decimal = $5,
    outstanding_balance = $6,
    interest_rate = $7,
    min_payment = $8,
    due_date = $9,
    notes = $10,
    payment_due_day = $11,
    monthly_obligation = $12,
    minimum_obligation = $13,
    remaining_months = $14,
    due_day = $15,
    is_interest_bearing = $16,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateDebtBalance :exec
UPDATE debts
SET outstanding_balance = $2,
    is_overdue = $3,
    updated_at = now()
WHERE id = $1;

-- name: UpdateDebtFromStatementParse :exec
UPDATE debts
SET
    min_payment = CASE WHEN sqlc.arg('min_payment') > 0 THEN sqlc.arg('min_payment') ELSE min_payment END,
    minimum_obligation = CASE WHEN sqlc.arg('min_payment') > 0 THEN sqlc.arg('min_payment') ELSE minimum_obligation END,
    principal_decimal = CASE WHEN sqlc.arg('principal_decimal') > 0 THEN sqlc.arg('principal_decimal') ELSE principal_decimal END,
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    updated_at = now()
WHERE id = sqlc.arg('id');

-- name: UpdateDebtHealthScore :exec
UPDATE debts
SET debt_health_score = $2, updated_at = now()
WHERE id = $1;

-- name: MarkDebtOverdue :exec
UPDATE debts
SET is_overdue = TRUE, updated_at = now()
WHERE id = $1;

-- name: AccrueInterest :exec
UPDATE debts
SET accrued_interest = accrued_interest + $2,
    last_accrual_date = $3,
    updated_at = now()
WHERE id = $1;

-- name: ResetAccruedInterest :exec
UPDATE debts
SET accrued_interest = 0,
    updated_at = now()
WHERE id = $1;

-- name: GetDebtForUpdate :one
SELECT *
FROM debts
WHERE id = $1 AND user_id = $2
FOR UPDATE;

-- name: UpdateDebtAfterPayment :exec
UPDATE debts
SET outstanding_balance = $2,
    accrued_interest = $3,
    last_accrual_date = $4,
    updated_at = now()
WHERE id = $1;

-- name: ListAllActiveDebts :many
SELECT *
FROM debts
WHERE outstanding_balance > 0 OR accrued_interest > 0;

-- name: DecrementRemainingMonthsAfterEMI :exec
UPDATE debts
SET remaining_months = CASE
        WHEN remaining_months IS NOT NULL AND remaining_months > 0 THEN remaining_months - 1
        ELSE remaining_months
    END,
    updated_at = now()
WHERE id = $1;

-- name: ExitMoratorium :exec
UPDATE debts
SET in_moratorium = FALSE,
    monthly_obligation = $2,
    updated_at = now()
WHERE id = $1;

-- name: ListLoansNeedingEMI :many
SELECT *
FROM debts
WHERE type NOT IN ('credit_card', 'bnpl')
  AND interest_model IS DISTINCT FROM 'revolving'::interest_model
  AND payment_due_day IS NOT NULL
  AND outstanding_balance > 0
  AND EXTRACT(DAY FROM CURRENT_DATE) = payment_due_day;

-- name: GetDebtBreakdownByType :many
-- Breakdown by legacy debt_type label (API compatibility).
SELECT
    type,
    COUNT(*) as count,
    COALESCE(SUM(outstanding_balance + accrued_interest), 0) as total_outstanding,
    COALESCE(SUM(minimum_obligation), 0) as total_min_payment,
    CASE
        WHEN SUM(outstanding_balance) > 0
        THEN ROUND((SUM(interest_rate * outstanding_balance) / SUM(outstanding_balance))::numeric, 5)
        ELSE 0
    END as avg_interest_rate
FROM debts
WHERE user_id = $1
    AND outstanding_balance > 0
GROUP BY type
ORDER BY total_outstanding DESC;

-- name: GetDebtsByTreatmentClassification :many
SELECT
    treatment_classification,
    COUNT(*) as count,
    COALESCE(SUM(outstanding_balance + accrued_interest), 0) as total_outstanding,
    COALESCE(AVG(debt_health_score), 0) as avg_health_score
FROM debts
WHERE user_id = $1
    AND outstanding_balance > 0
    AND treatment_classification IS NOT NULL
GROUP BY treatment_classification
ORDER BY treatment_classification;
