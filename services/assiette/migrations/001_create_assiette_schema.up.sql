CREATE SCHEMA assiette;

CREATE TABLE assiette.state (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    data JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assiette.custom_recipes (
    id TEXT PRIMARY KEY,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
