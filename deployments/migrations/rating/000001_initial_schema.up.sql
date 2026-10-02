CREATE TABLE IF NOT EXISTS feedbacks (
    id UUID PRIMARY KEY,
    ride_id UUID NOT NULL,
    reviewer_id UUID NOT NULL,
    reviewee_id UUID NOT NULL,
    score NUMERIC(3, 2) NOT NULL,
    comment TEXT,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedbacks_reviewee ON feedbacks(reviewee_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_ride ON feedbacks(ride_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_reviewer ON feedbacks(reviewer_id);
