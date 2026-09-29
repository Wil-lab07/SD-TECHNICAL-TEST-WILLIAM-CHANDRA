CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    name VARCHAR(100) NOT NULL,
    height DECIMAL(5,2) NOT NULL,
    weight DECIMAL(5,2) NOT NULL,
    position VARCHAR(20) NOT NULL CHECK (position IN ('striker', 'midfielder', 'defender', 'goalkeeper')),
    jersey_number INT NOT NULL CHECK (jersey_number > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX idx_players_team_id_jersey_number_unique 
ON players (team_id, jersey_number) 
WHERE deleted_at IS NULL;
