CREATE TABLE assiette.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    lang TEXT NOT NULL DEFAULT 'fr' CHECK (lang IN ('fr', 'en')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assiette.sessions (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES assiette.users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON assiette.sessions(user_id);

CREATE TABLE assiette.teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assiette.team_members (
    user_id UUID PRIMARY KEY REFERENCES assiette.users(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES assiette.teams(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_team_members_team_id ON assiette.team_members(team_id);

CREATE TABLE assiette.team_state (
    team_id UUID PRIMARY KEY REFERENCES assiette.teams(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assiette.invites (
    token_hash TEXT PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES assiette.teams(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES assiette.users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE assiette.custom_recipes ADD COLUMN team_id UUID REFERENCES assiette.teams(id) ON DELETE CASCADE;
ALTER TABLE assiette.custom_recipes DROP CONSTRAINT custom_recipes_pkey;
ALTER TABLE assiette.custom_recipes ADD CONSTRAINT custom_recipes_team_id_id_key UNIQUE (team_id, id);
