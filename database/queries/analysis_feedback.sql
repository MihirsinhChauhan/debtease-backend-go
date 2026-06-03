-- name: CreateAnalysisFeedback :one
INSERT INTO analysis_feedback (
    user_id,
    session_id,
    seq_score,
    trust_score,
    accuracy_feedback,
    feedback_text,
    screen_name,
    context_metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAnalysisFeedback :many
SELECT
    af.id,
    af.created_at,
    af.user_id,
    u.email AS user_email,
    af.session_id,
    af.screen_name,
    af.trust_score,
    af.seq_score,
    af.accuracy_feedback,
    af.feedback_text,
    af.context_metadata
FROM analysis_feedback af
LEFT JOIN users u ON u.id = af.user_id
WHERE (sqlc.narg('since')::timestamptz IS NULL OR af.created_at >= sqlc.narg('since'))
  AND (sqlc.narg('screen_name')::text IS NULL OR af.screen_name = sqlc.narg('screen_name'))
  AND (
    sqlc.narg('priority_only')::boolean IS NOT TRUE
    OR af.trust_score < 4
    OR af.accuracy_feedback = 'No'
  )
  AND (sqlc.narg('trust_lt')::smallint IS NULL OR af.trust_score < sqlc.narg('trust_lt'))
  AND (sqlc.narg('accuracy')::text IS NULL OR af.accuracy_feedback = sqlc.narg('accuracy'))
ORDER BY af.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAnalysisFeedback :one
SELECT COUNT(*)::bigint AS count
FROM analysis_feedback af
WHERE (sqlc.narg('since')::timestamptz IS NULL OR af.created_at >= sqlc.narg('since'))
  AND (sqlc.narg('screen_name')::text IS NULL OR af.screen_name = sqlc.narg('screen_name'))
  AND (
    sqlc.narg('priority_only')::boolean IS NOT TRUE
    OR af.trust_score < 4
    OR af.accuracy_feedback = 'No'
  )
  AND (sqlc.narg('trust_lt')::smallint IS NULL OR af.trust_score < sqlc.narg('trust_lt'))
  AND (sqlc.narg('accuracy')::text IS NULL OR af.accuracy_feedback = sqlc.narg('accuracy'));

-- name: FeedbackSummaryTotals :one
SELECT
    COUNT(*)::bigint AS total,
    COALESCE(AVG(trust_score), 0)::float8 AS avg_trust,
    COALESCE(AVG(seq_score), 0)::float8 AS avg_seq
FROM analysis_feedback
WHERE (sqlc.narg('since')::timestamptz IS NULL OR created_at >= sqlc.narg('since'));

-- name: FeedbackSummaryAccuracy :many
SELECT accuracy_feedback, COUNT(*)::bigint AS count
FROM analysis_feedback
WHERE (sqlc.narg('since')::timestamptz IS NULL OR created_at >= sqlc.narg('since'))
  AND accuracy_feedback IS NOT NULL
GROUP BY accuracy_feedback
ORDER BY count DESC;

-- name: FeedbackSummaryByScreen :many
SELECT screen_name, COUNT(*)::bigint AS count
FROM analysis_feedback
WHERE (sqlc.narg('since')::timestamptz IS NULL OR created_at >= sqlc.narg('since'))
GROUP BY screen_name
ORDER BY count DESC;

-- name: FeedbackTrustByScoreBand :many
SELECT
    context_metadata->>'score_band' AS metadata_value,
    AVG(trust_score)::float8 AS avg_trust,
    COUNT(*)::bigint AS count
FROM analysis_feedback
WHERE (sqlc.narg('since')::timestamptz IS NULL OR created_at >= sqlc.narg('since'))
  AND trust_score IS NOT NULL
  AND context_metadata->>'score_band' IS NOT NULL
  AND context_metadata->>'score_band' != ''
GROUP BY context_metadata->>'score_band'
ORDER BY count DESC;

-- name: FeedbackTrustByHealthStatus :many
SELECT
    COALESCE(
        context_metadata->>'health_status',
        context_metadata->>'health_status_after'
    ) AS metadata_value,
    AVG(trust_score)::float8 AS avg_trust,
    COUNT(*)::bigint AS count
FROM analysis_feedback
WHERE (sqlc.narg('since')::timestamptz IS NULL OR created_at >= sqlc.narg('since'))
  AND trust_score IS NOT NULL
  AND COALESCE(
        context_metadata->>'health_status',
        context_metadata->>'health_status_after'
    ) IS NOT NULL
  AND COALESCE(
        context_metadata->>'health_status',
        context_metadata->>'health_status_after'
    ) != ''
GROUP BY COALESCE(
    context_metadata->>'health_status',
    context_metadata->>'health_status_after'
)
ORDER BY count DESC;

-- name: ListFeedbackByUserID :many
SELECT
    af.id,
    af.created_at,
    af.user_id,
    u.email AS user_email,
    af.session_id,
    af.screen_name,
    af.trust_score,
    af.seq_score,
    af.accuracy_feedback,
    af.feedback_text,
    af.context_metadata
FROM analysis_feedback af
LEFT JOIN users u ON u.id = af.user_id
WHERE af.user_id = $1
ORDER BY af.created_at DESC
LIMIT $2;

-- name: ListFeedbackBySessionID :many
SELECT
    af.id,
    af.created_at,
    af.user_id,
    u.email AS user_email,
    af.session_id,
    af.screen_name,
    af.trust_score,
    af.seq_score,
    af.accuracy_feedback,
    af.feedback_text,
    af.context_metadata
FROM analysis_feedback af
LEFT JOIN users u ON u.id = af.user_id
WHERE af.session_id = $1
ORDER BY af.created_at DESC
LIMIT $2;
