-- name: InsertFinancialSnapshot :one
INSERT INTO financial_snapshots (
    user_id,
    session_id,
    snapshot_type,
    snapshot_source,
    monthly_income,
    fixed_expenses,
    variable_expenses,
    savings_cash,
    investments_value,
    total_monthly_emi,
    other_income,
    total_debt_balance,
    trigger_event,
    extra_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: GetLatestSnapshotBySession :one
SELECT *
FROM financial_snapshots
WHERE session_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLatestSnapshotByUser :one
SELECT *
FROM financial_snapshots
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;
