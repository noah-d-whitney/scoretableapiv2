CREATE VIEW games_view AS
SELECT p.id AS pin_id, p.pin, p.scope, g.id, g.user_id, g.created_at, g.version, g.status,
       g.date_time, g.team_size, g.period_length, g.period_count, g.score_target, (
            CASE WHEN g.score_target IS NULL AND g.period_count IS NULL
                    THEN 'manual'
                WHEN g.score_target IS NOT NULL AND g.period_count IS NULL
                    THEN 'target'
                WHEN g.score_target IS NULL AND g.period_count IS NOT NULL
                    THEN 'timed'
            END
            ) AS type, (
            SELECT p.pin
                FROM pins p
                    JOIN public.teams t on p.id = t.pin_id
                    JOIN public.games_teams gt on t.id = gt.team_id
                WHERE gt.game_id = g.id AND gt.side = 0
            ) AS home_team_pin, (
            SELECT p.pin
                FROM pins p
                    JOIN public.teams t on p.id = t.pin_id
                    JOIN public.games_teams gt on t.id = gt.team_id
                WHERE gt.game_id = g.id AND gt.side = 1
            ) AS away_team_pin, ARRAY(
            SELECT pin
                FROM pins
                    JOIN public.players p2 on pins.id = p2.pin_id
                    JOIN public.teams_players t on p2.id = t.player_id
                    JOIN public.games_teams gt on t.team_id = gt.team_id
                WHERE gt.game_id = g.id
            ) AS player_pins, ARRAY(
            SELECT pin
                FROM pins
                    JOIN public.players p3 on pins.id = p3.pin_id
                    JOIN public.teams_players tp on p3.id = tp.player_id
                    JOIN public.games_teams gt2 on tp.team_id = gt2.team_id
                WHERE gt2.game_id = g.id AND gt2.side = 0
            ) AS home_player_pins, ARRAY(
            SELECT pin
                FROM pins
                    JOIN public.players p3 on pins.id = p3.pin_id
                    JOIN public.teams_players tp on p3.id = tp.player_id
                    JOIN public.games_teams gt2 on tp.team_id = gt2.team_id
                WHERE gt2.game_id = g.id AND gt2.side = 1
            ) AS away_player_pins
FROM games g
    JOIN public.pins p on p.id = g.pin_id