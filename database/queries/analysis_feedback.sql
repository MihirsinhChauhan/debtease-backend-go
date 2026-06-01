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
