CREATE TABLE IF NOT EXISTS match_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id),
    player_id UUID NOT NULL REFERENCES players(id),
    team_id UUID NOT NULL REFERENCES teams(id),
    minute INT NOT NULL CHECK (minute BETWEEN 1 AND 120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
