CREATE INDEX IF NOT EXISTS idx_revoked_tokens_expires_at ON revoked_tokens (expires_at);

CREATE INDEX IF NOT EXISTS idx_matches_home_team_finished ON matches (home_team_id, match_date) 
  WHERE deleted_at IS NULL AND status = 'finished';

CREATE INDEX IF NOT EXISTS idx_matches_away_team_finished ON matches (away_team_id, match_date) 
  WHERE deleted_at IS NULL AND status = 'finished';

CREATE INDEX IF NOT EXISTS idx_match_goals_match_id ON match_goals (match_id) 
  WHERE deleted_at IS NULL;
