CREATE TABLE assiette.recipe_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES assiette.teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES assiette.users(id) ON DELETE CASCADE,
    recipe_id TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recipe_comments_team_recipe ON assiette.recipe_comments(team_id, recipe_id, created_at);
