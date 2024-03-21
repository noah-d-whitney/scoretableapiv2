ALTER TABLE IF EXISTS teams_players
DROP COLUMN IF EXISTS player_number,
DROP COLUMN IF EXISTS lineup_number,
DROP CONSTRAINT IF EXISTS teams_players_player_number_unq,
DROP CONSTRAINT IF EXISTS teams_players_lineup_number_unq;