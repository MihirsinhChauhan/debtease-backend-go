-- name: CreateDebt :one
INSERT INTO debts (
  user_id, name, type, lender, principal_decimal,
  outstanding_balance, interest_rate, min_payment,
  due_date, payment_frequency, debt_taken,
  risk_level, priority, notes
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING id, user_id, name, type, lender, principal_decimal, outstanding_balance, interest_rate, min_payment, due_date, payment_frequency, is_overdue, late_fees, other_charges, debt_taken, risk_level, debt_health_score, priority, notes, created_at, updated_at;

-- name: ListUserDebts :many
SELECT id, user_id, name, type, lender, principal_decimal, outstanding_balance, interest_rate, min_payment, due_date, payment_frequency, is_overdue, late_fees, other_charges, debt_taken, risk_level, debt_health_score, priority, notes, created_at, updated_at
FROM debts
WHERE user_id = $1
ORDER BY priority DESC, due_date ASC;

-- name: GetDebtByID :one
SELECT id, user_id, name, type, lender, principal_decimal, outstanding_balance, interest_rate, min_payment, due_date, payment_frequency, is_overdue, late_fees, other_charges, debt_taken, risk_level, debt_health_score, priority, notes, created_at, updated_at
FROM debts
WHERE id = $1 AND user_id = $2;

-- name: UpdateDebtBalance :exec
UPDATE debts
SET outstanding_balance = $2,
    is_overdue = $3,
    updated_at = now()
WHERE id = $1;

-- name: UpdateDebtHealthScore :exec
UPDATE debts
SET debt_health_score = $2, updated_at = now()
WHERE id = $1;

-- name: MarkDebtOverdue :exec
UPDATE debts
SET is_overdue = TRUE, updated_at = now()
WHERE id = $1;
