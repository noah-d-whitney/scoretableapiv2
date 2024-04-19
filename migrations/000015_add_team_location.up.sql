ALTER TABLE IF EXISTS teams
    DROP CONSTRAINT unq_userid_team_name;

ALTER TABLE IF EXISTS teams
    ADD COLUMN location text DEFAULT null,
    ADD CONSTRAINT unq_userid_team_name 
        unique (user_id, name, location);
