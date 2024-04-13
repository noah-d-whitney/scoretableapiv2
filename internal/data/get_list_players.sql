SELECT pins.id, pins.pin, pins.scope, players.id, players.first_name, players.last_name, players.pref_number,
            players.created_at, players.version, (
				SELECT count(*)::int::bool
					FROM teams_players
					WHERE player_id = players.id)
        FROM players
        INNER JOIN pins ON players.pin_id = pins.pin_id
        WHERE players.user_id = $1 AND pins.pin = ANY($2)`
