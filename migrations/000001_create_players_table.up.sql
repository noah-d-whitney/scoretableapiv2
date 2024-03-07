CREATE TABLE IF NOT EXISTS players (
    id bigserial PRIMARY KEY,
    first_name text NOT NULL,
    last_name text NOT NULL,
    pref_number integer NOT NULL,
    is_active bool NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);

ALTER TABLE players ADD CONSTRAINT players_pref_number_check CHECK ( pref_number >= 0 AND
                                                                     pref_number < 100 );
