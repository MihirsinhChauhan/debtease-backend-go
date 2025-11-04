-- name: RecordPayment :one
INSERT INTO payments_history (
  debt_id, amount, payment_date, type,
  principal_applied, interest_applied,
  notes, payment_method
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListPaymentsByDebt :many
SELECT *
FROM payments_history
WHERE debt_id = $1
ORDER BY payment_date DESC
LIMIT $2 OFFSET $3;

-- name: GetPaymentByID :one
SELECT *
FROM payments_history
WHERE id = $1 AND debt_id = $2;

-- name: GetRecentPayments :many
SELECT p.*, d.name AS debt_name
FROM payments_history p
JOIN debts d ON p.debt_id = d.id
WHERE p.payment_date >= $1
  AND d.user_id = $2
ORDER BY p.payment_date DESC;