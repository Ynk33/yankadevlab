package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
)

const (
	maxRequestRunes = 1000
	maxRecipes      = 5
	draftTTL        = "7 days"
	generateTimeout = 3 * time.Minute
)

type generateInput struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
}

type generateOutput struct {
	Matches   []string `json:"matches"`
	Generated []Recipe `json:"generated"`
}

// ListCatalog returns the generated recipes users added to the global catalog, oldest first.
func (h *StoreHandler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	var data []byte
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT COALESCE(jsonb_agg(data ORDER BY added_at), '[]'::jsonb) FROM assiette.catalog_recipes WHERE added_at IS NOT NULL`).Scan(&data)
	if err != nil {
		h.internalError(w, "failed to list catalog recipes", err)
		return
	}
	writeJSON(w, data)
}

// AddToCatalog publishes a generated draft to the global catalog. Only drafts written by /api/generate can be
// published, so the catalog never holds client-supplied recipe content.
func (h *StoreHandler) AddToCatalog(w http.ResponseWriter, r *http.Request) {
	res, err := h.DB.ExecContext(r.Context(),
		`UPDATE assiette.catalog_recipes SET added_at = COALESCE(added_at, now()) WHERE id = $1`, chi.URLParam(r, "id"))
	if err != nil {
		h.internalError(w, "failed to add catalog recipe", err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, `{"error":"unknown or expired recipe"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Generate answers a free-text request with n recipes: catalog recipes that fulfil it (score >= MatchThreshold)
// first, then generated ones, stored as drafts until a user adds them to the catalog.
func (h *StoreHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if h.Gen == nil {
		http.Error(w, `{"error":"generation is not configured"}`, http.StatusServiceUnavailable)
		return
	}
	var in generateInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&in); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	in.Prompt = strings.TrimSpace(in.Prompt)
	if in.Prompt == "" || utf8.RuneCountInString(in.Prompt) > maxRequestRunes || in.N < 1 || in.N > maxRecipes {
		http.Error(w, `{"error":"prompt must be 1 to 1000 characters and n between 1 and 5"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), generateTimeout)
	defer cancel()
	disliked, exclusions, err := h.teamFilters(ctx)
	if err != nil {
		h.internalError(w, "failed to read team state", err)
		return
	}
	published, err := h.publishedRecipes(ctx)
	if err != nil {
		h.internalError(w, "failed to read catalog recipes", err)
		return
	}
	request := in.Prompt
	if exclusions != "" {
		request += "\nNever use: " + exclusions
	}
	extra := make([]catalogEntry, len(published))
	for i, p := range published {
		extra[i] = recipeEntry(p)
	}

	matches, err := h.Gen.Match(ctx, request, extra, disliked, in.N)
	if err != nil {
		h.generationError(w, err)
		return
	}
	out := generateOutput{Matches: matches, Generated: []Recipe{}}
	if need := in.N - len(matches); need > 0 {
		recipes, err := h.Gen.Generate(ctx, request, h.titles(matches, extra), need, time.Now().Month())
		if err != nil {
			h.generationError(w, err)
			return
		}
		if err := h.saveDrafts(ctx, recipes, userIDFrom(r.Context())); err != nil {
			h.internalError(w, "failed to save generated recipes", err)
			return
		}
		out.Generated = recipes
	}
	data, err := json.Marshal(out)
	if err != nil {
		h.internalError(w, "failed to encode generated recipes", err)
		return
	}
	writeJSON(w, data)
}

func (h *StoreHandler) generationError(w http.ResponseWriter, err error) {
	if errors.Is(err, errRefused) {
		http.Error(w, `{"error":"request declined"}`, http.StatusUnprocessableEntity)
		return
	}
	h.Log.Error("recipe generation failed", "error", err)
	http.Error(w, `{"error":"generation failed"}`, http.StatusBadGateway)
}

// teamFilters reads the team's disliked recipe ids and comma-separated exclusions from its state.
func (h *StoreHandler) teamFilters(ctx context.Context) (map[string]bool, string, error) {
	var raw []byte
	err := h.DB.QueryRowContext(ctx, `SELECT data FROM assiette.team_state WHERE team_id = $1`, teamIDFrom(ctx)).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]bool{}, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var state struct {
		Dislikes map[string]json.RawMessage `json:"dislikes"`
		Prefs    struct {
			Exclusions string `json:"exclusions"`
		} `json:"prefs"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return map[string]bool{}, "", nil
	}
	disliked := make(map[string]bool, len(state.Dislikes))
	for id := range state.Dislikes {
		disliked[id] = true
	}
	return disliked, strings.TrimSpace(state.Prefs.Exclusions), nil
}

func (h *StoreHandler) publishedRecipes(ctx context.Context) ([]Recipe, error) {
	var raw []byte
	err := h.DB.QueryRowContext(ctx,
		`SELECT COALESCE(jsonb_agg(data ORDER BY added_at), '[]'::jsonb) FROM assiette.catalog_recipes WHERE added_at IS NOT NULL`).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var recipes []Recipe
	return recipes, json.Unmarshal(raw, &recipes)
}

// titles returns the titles of the matched recipes, so the generator writes different ones.
func (h *StoreHandler) titles(ids []string, extra []catalogEntry) []string {
	var out []string
	for _, e := range slices.Concat(h.Gen.catalog, extra) {
		if slices.Contains(ids, e.ID) {
			out = append(out, e.Title)
		}
	}
	return out
}

// saveDrafts stores generated recipes as unpublished drafts and purges drafts older than draftTTL.
func (h *StoreHandler) saveDrafts(ctx context.Context, recipes []Recipe, userID string) error {
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM assiette.catalog_recipes WHERE added_at IS NULL AND created_at < now() - $1::interval`, draftTTL); err != nil {
		return err
	}
	for _, recipe := range recipes {
		data, err := json.Marshal(recipe)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO assiette.catalog_recipes (id, data, created_by) VALUES ($1, $2, $3)`, recipe.ID, data, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
