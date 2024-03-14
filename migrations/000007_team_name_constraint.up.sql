ALTER TABLE teams ADD CONSTRAINT unq_userid_team_name
    unique (user_id, name);