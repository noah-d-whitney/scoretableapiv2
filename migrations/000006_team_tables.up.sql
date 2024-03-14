CREATE TABLE IF NOT EXISTS pins (
    id bigserial PRIMARY KEY,
    pin text NOT NULL UNIQUE,
    scope text NOT NULL
);

CREATE TABLE IF NOT EXISTS teams (
    id bigserial PRIMARY KEY,
    pin_id bigint NOT NULL REFERENCES pins ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    name text NOT NULL,
    version integer NOT NULL DEFAULT 1,
    size integer NOT NULL DEFAULT 0,
    is_active bool NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS teams_players (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    team_id bigint NOT NULL REFERENCES teams ON DELETE CASCADE,
    player_id bigint NOT NULL REFERENCES players ON DELETE CASCADE,
    PRIMARY KEY (team_id, player_id, user_id)
);