CREATE TABLE IF NOT EXISTS games (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    pin_id bigint NOT NULL REFERENCES pins ON DELETE NO ACTION,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    status integer NOT NULL DEFAULT 0,
    date_time timestamp(0) WITH TIME ZONE,
    team_size integer NOT NULL,
    period_length integer,
    period_count integer,
    score_target integer
);

CREATE TABLE IF NOT EXISTS games_teams (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    game_id bigint NOT NULL REFERENCES games ON DELETE CASCADE,
    team_id bigint NOT NULL REFERENCES teams ON DELETE CASCADE,
    PRIMARY KEY (user_id, game_id, team_id)
);