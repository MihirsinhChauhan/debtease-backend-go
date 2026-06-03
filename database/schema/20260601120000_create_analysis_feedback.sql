-- +goose Up
-- +goose StatementBegin

CREATE TABLE analysis_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID NULL REFERENCES calculation_sessions(session_id) ON DELETE SET NULL,
    seq_score SMALLINT NULL CHECK (seq_score BETWEEN 1 AND 7),
    trust_score SMALLINT NULL CHECK (trust_score BETWEEN 1 AND 7),
    accuracy_feedback TEXT NULL CHECK (accuracy_feedback IN ('Yes', 'Mostly', 'No')),
    feedback_text TEXT NULL,
    screen_name TEXT NOT NULL,
    context_metadata JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_analysis_feedback_user_id ON analysis_feedback(user_id);
CREATE INDEX idx_analysis_feedback_session_id ON analysis_feedback(session_id);
CREATE INDEX idx_analysis_feedback_screen_name ON analysis_feedback(screen_name);
CREATE INDEX idx_analysis_feedback_trust ON analysis_feedback(trust_score);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS analysis_feedback;

-- +goose StatementEnd
