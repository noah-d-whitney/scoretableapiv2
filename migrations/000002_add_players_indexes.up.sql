CREATE INDEX IF NOT EXISTS players_first_name_idx ON players USING gin (to_tsvector('simple',
                                                                                    first_name));
CREATE INDEX IF NOT EXISTS players_last_name_idx ON players USING gin (to_tsvector('simple',
                                                                                   last_name));