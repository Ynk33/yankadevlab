CREATE TABLE assiette.catalog_recipes (
    id TEXT PRIMARY KEY,
    data JSONB NOT NULL,
    created_by UUID REFERENCES assiette.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_at TIMESTAMPTZ
);

CREATE INDEX idx_catalog_recipes_added ON assiette.catalog_recipes(added_at) WHERE added_at IS NOT NULL;
