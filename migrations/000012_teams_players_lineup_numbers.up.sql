ALTER TABLE IF EXISTS teams_players
ADD COLUMN player_number integer,
ADD COLUMN lineup_number integer,
ADD CONSTRAINT teams_players_player_number_unq UNIQUE(player_number, team_id),
ADD CONSTRAINT teams_players_lineup_number_unq UNIQUE(lineup_number, team_id);
