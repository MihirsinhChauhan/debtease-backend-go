-- name: InsertFinancialEvent :one
INSERT INTO financial_events (
    user_id,
    event_type,
    entity_type,
    entity_id,
    amount,
    currency,
    occurred_at,
    metadata,
    schema_version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListFinancialEventsByUser :many
SELECT *
FROM financial_events
WHERE user_id = $1
ORDER BY occurred_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: ListFinancialEventsForDebt :many
SELECT *
FROM financial_events
WHERE user_id = $1
  AND entity_type = 'debt'
  AND entity_id = $2
ORDER BY occurred_at DESC, id DESC
LIMIT $3 OFFSET $4;
