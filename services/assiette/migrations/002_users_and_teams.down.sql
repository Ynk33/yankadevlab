DELETE FROM assiette.custom_recipes a USING assiette.custom_recipes b
    WHERE a.id = b.id AND a.ctid > b.ctid;
ALTER TABLE assiette.custom_recipes DROP CONSTRAINT custom_recipes_team_id_id_key;
ALTER TABLE assiette.custom_recipes DROP COLUMN team_id;
ALTER TABLE assiette.custom_recipes ADD PRIMARY KEY (id);

DROP TABLE assiette.invites;
DROP TABLE assiette.team_state;
DROP TABLE assiette.team_members;
DROP TABLE assiette.teams;
DROP TABLE assiette.sessions;
DROP TABLE assiette.users;
