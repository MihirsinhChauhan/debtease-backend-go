-- name: CreateStatementUpload :one
INSERT INTO statement_uploads (
    debt_id, storage_path, original_filename, checksum, mime, status
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateStatementUploadStatus :one
UPDATE statement_uploads
SET status = $2,
    failure_code = $3,
    error_message = $4,
    statement_id = $5
WHERE id = $1
RETURNING *;

-- name: GetStatementUploadByID :one
SELECT su.*
FROM statement_uploads su
JOIN debts d ON d.id = su.debt_id
WHERE su.id = $1 AND d.user_id = $2;

-- name: GetStatementUploadByIDForDebt :one
SELECT * FROM statement_uploads
WHERE id = $1 AND debt_id = $2;

-- name: CreateCreditCardStatement :one
INSERT INTO credit_card_statements (
    debt_id, statement_start, statement_end, statement_date, due_date,
    statement_amount, minimum_due, previous_balance, payments_received,
    new_balance, credit_limit, available_credit, source, upload_id,
    issuer_detected, parser_variant, metadata
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13, $14,
    $15, $16, $17
)
RETURNING *;

-- name: GetCreditCardStatementByID :one
SELECT cs.*
FROM credit_card_statements cs
JOIN debts d ON d.id = cs.debt_id
WHERE cs.id = $1 AND d.user_id = $2;

-- name: ListCreditCardStatementsByDebtID :many
SELECT cs.*
FROM credit_card_statements cs
JOIN debts d ON d.id = cs.debt_id
WHERE cs.debt_id = $1 AND d.user_id = $2
ORDER BY cs.statement_date DESC NULLS LAST, cs.created_at DESC;

-- name: CountCreditCardStatementsByDebtID :one
SELECT COUNT(*)::bigint AS count
FROM credit_card_statements
WHERE debt_id = $1;
