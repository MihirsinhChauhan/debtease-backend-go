-- name: CountSessionsSince :one
SELECT COUNT(*)::bigint AS count
FROM calculation_sessions
WHERE created_at >= $1;

-- name: CountOnboardingSnapshotsSince :one
SELECT COUNT(*)::bigint AS count
FROM financial_snapshots
WHERE snapshot_type = 'onboarding'
  AND created_at >= $1;

-- name: CountSignupsSince :one
SELECT COUNT(*)::bigint AS count
FROM users
WHERE created_at >= $1;

-- name: CountFeedbackSince :one
SELECT COUNT(*)::bigint AS count
FROM analysis_feedback
WHERE created_at >= $1;

-- name: AvgFeedbackTrustSince :one
SELECT AVG(trust_score)::float8 AS avg
FROM analysis_feedback
WHERE created_at >= $1 AND trust_score IS NOT NULL;

-- name: AvgFeedbackSeqSince :one
SELECT AVG(seq_score)::float8 AS avg
FROM analysis_feedback
WHERE created_at >= $1 AND seq_score IS NOT NULL;

-- name: FeedbackAccuracyDistributionSince :many
SELECT accuracy_feedback, COUNT(*)::bigint AS count
FROM analysis_feedback
WHERE created_at >= $1
  AND accuracy_feedback IS NOT NULL
GROUP BY accuracy_feedback
ORDER BY count DESC;
