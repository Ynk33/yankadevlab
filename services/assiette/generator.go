package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

const (
	matchModel    = "claude-haiku-5-5"
	generateModel = "claude-sonnet-5-5"
	// MatchThreshold is the score (0-100) from which a catalog recipe is reused instead of generating one.
	MatchThreshold   = 80
	maxMatches       = 10
	requestTimeout   = 2 * time.Minute
	generateAttempts = 2
)

var errRefused = errors.New("request declined by the model")

const matchInstructions = `You help users of Assiette, a French home-cooking app, find recipes. You receive the catalog (one recipe per line: id | title | protein | time | tags | ingredients) and a request describing what the user feels like eating, what they have at hand, or constraints. Score how well each recipe fulfils the whole request, from 0 (unrelated) to 100 (exactly what was asked): ingredients they want to use, cravings, time, diet and exclusions all count. A recipe that ignores an explicit constraint scores below 50. Return at most 10 recipes scoring 50 or more, best first, with their exact id. Return an empty list when nothing fits.

Catalog:
`

const generateInstructions = `You are the chef of Assiette, a French home-cooking app for healthy everyday lunches and dinners. Write recipes that make people want to cook tonight: a precise, appetising French title (the dish's real name when it has one, never generic like "Poulet aux légumes"), a clear idea of what makes the dish good (a texture, a sauce, a contrast, a technique), and seasonal produce. Each recipe answers the user's request.

Rules:
- A complete main dish: plenty of vegetables, one main protein (meat, fish, seafood, eggs, tofu or tempeh, legumes or cheese), a wholegrain starch or legumes, a good fat. No cured meats, no ultra-processed products.
- Ingredients come only from the vocabulary below; amounts are for 2 people, in the unit given for each ingredient. Leave out salt, pepper and water.
- 3 to 6 French imperative steps, specific (temperatures, times, visual cues), in the order they are done. Then the same steps translated into natural English, and the title translated into English.
- minutes is the total active and cooking time. meal: "midi" for light, cold or quick dishes, "soir" for hot cooked dishes, "both" when either fits. world: true for a cuisine from outside Europe. batch: true when it is cooked in quantity and keeps 3 to 4 days.
- The recipes differ from each other and from the ones listed as already suggested.

Vocabulary (name | unit | aisle):
`

// Generator calls Claude to match catalog recipes against a request and to write new recipes.
type Generator struct {
	client     anthropic.Client
	products   map[string]Product
	names      []string
	catalog    []catalogEntry
	matchBlock string
	vocabBlock string
	log        *slog.Logger
}

// NewGenerator builds a generator from the embedded product table and catalog; the API key comes from the environment.
func NewGenerator(apiKey string, log *slog.Logger) (*Generator, error) {
	products, err := loadProducts()
	if err != nil {
		return nil, err
	}
	catalog, err := staticCatalog()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(products))
	for name := range products {
		names = append(names, name)
	}
	slices.Sort(names)
	lines := make([]string, len(catalog))
	for i, e := range catalog {
		lines[i] = e.line()
	}
	vocab := make([]string, len(names))
	for i, name := range names {
		vocab[i] = fmt.Sprintf("%s | %s | %s", name, products[name].Unit, products[name].Rayon)
	}
	return &Generator{
		client:     anthropic.NewClient(option.WithAPIKey(apiKey), option.WithRequestTimeout(requestTimeout)),
		products:   products,
		names:      names,
		catalog:    catalog,
		matchBlock: matchInstructions + strings.Join(lines, "\n"),
		vocabBlock: generateInstructions + strings.Join(vocab, "\n"),
		log:        log,
	}, nil
}

type match struct {
	ID    string `json:"id"`
	Score int    `json:"score"`
}

// Match asks the fast model which catalog recipes fulfil the request; extra holds recipes added to the global
// catalog, excluded removes ids (disliked recipes). It returns up to n ids scoring at least MatchThreshold.
func (g *Generator) Match(ctx context.Context, request string, extra []catalogEntry, excluded map[string]bool, n int) ([]string, error) {
	var user strings.Builder
	if len(extra) > 0 {
		user.WriteString("Also in the catalog:\n")
		for _, e := range extra {
			user.WriteString(e.line() + "\n")
		}
		user.WriteString("\n")
	}
	user.WriteString("Request: " + request)
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"matches": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"properties":           map[string]any{"id": map[string]any{"type": "string"}, "score": map[string]any{"type": "integer"}},
					"required":             []string{"id", "score"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"matches"},
		"additionalProperties": false,
	}
	resp, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     matchModel,
		MaxTokens: 2000,
		System:    []anthropic.TextBlockParam{{Text: g.matchBlock, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
		Thinking:  anthropic.ThinkingConfigParamUnion{OfDisabled: &anthropic.ThinkingConfigDisabledParam{}},
		OutputConfig: anthropic.OutputConfigParam{
			Effort: anthropic.OutputConfigEffortLow,
			Format: anthropic.JSONOutputFormatParam{Schema: schema},
		},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(user.String()))},
	})
	if err != nil {
		return nil, fmt.Errorf("match: %w", err)
	}
	g.log.Info("catalog match", "input_tokens", resp.Usage.InputTokens, "cache_read", resp.Usage.CacheReadInputTokens, "output_tokens", resp.Usage.OutputTokens)
	if resp.StopReason == anthropic.StopReasonRefusal {
		return nil, errRefused
	}
	var out struct {
		Matches []match `json:"matches"`
	}
	if err := json.Unmarshal([]byte(textOf(resp.Content)), &out); err != nil {
		return nil, fmt.Errorf("match: parse answer: %w", err)
	}
	known := map[string]bool{}
	for _, e := range g.catalog {
		known[e.ID] = true
	}
	for _, e := range extra {
		known[e.ID] = true
	}
	slices.SortStableFunc(out.Matches, func(a, b match) int { return b.Score - a.Score })
	ids := []string{}
	for _, m := range out.Matches {
		if len(ids) == n {
			break
		}
		if m.Score >= MatchThreshold && known[m.ID] && !excluded[m.ID] && !slices.Contains(ids, m.ID) {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

// Generate asks the creative model for n new recipes answering the request, avoiding the listed titles, and keeps
// the cards that pass buildRecipe. Cards dropped by buildRecipe are asked for once more.
func (g *Generator) Generate(ctx context.Context, request string, avoid []string, n int, month time.Month) ([]Recipe, error) {
	var recipes []Recipe
	for attempt := 0; attempt < generateAttempts && len(recipes) < n; attempt++ {
		batch, err := g.generateOnce(ctx, request, avoid, n-len(recipes), month)
		if err != nil {
			return nil, err
		}
		for _, r := range batch {
			recipes = append(recipes, r)
			avoid = append(avoid, r.Title)
		}
	}
	return recipes, nil
}

func (g *Generator) generateOnce(ctx context.Context, request string, avoid []string, n int, month time.Month) ([]Recipe, error) {
	ingredient := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{"name": map[string]any{"type": "string", "enum": g.names}, "qty": map[string]any{"type": "number"}},
		"required":             []string{"name", "qty"},
		"additionalProperties": false,
	}
	card := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":       map[string]any{"type": "string"},
			"title_en":    map[string]any{"type": "string"},
			"meal":        map[string]any{"type": "string", "enum": []string{"midi", "soir", "both"}},
			"minutes":     map[string]any{"type": "integer"},
			"world":       map[string]any{"type": "boolean"},
			"batch":       map[string]any{"type": "boolean"},
			"ingredients": map[string]any{"type": "array", "items": ingredient},
			"steps":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"steps_en":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required":             []string{"title", "title_en", "meal", "minutes", "world", "batch", "ingredients", "steps", "steps_en"},
		"additionalProperties": false,
	}
	schema := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{"recipes": map[string]any{"type": "array", "items": card}},
		"required":             []string{"recipes"},
		"additionalProperties": false,
	}
	user := fmt.Sprintf("Month: %s.\nWrite %d recipe(s).\nRequest: %s", monthsFR[month-1], n, request)
	if len(avoid) > 0 {
		user += "\nAlready suggested: " + strings.Join(avoid, " ; ")
	}
	resp, err := g.client.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
		Model:     generateModel,
		MaxTokens: 16000,
		Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		Fallbacks: anthropic.BetaFallbacksParamOfDefault(),
		System:    []anthropic.BetaTextBlockParam{{Text: g.vocabBlock, CacheControl: anthropic.NewBetaCacheControlEphemeralParam()}},
		OutputConfig: anthropic.BetaOutputConfigParam{
			Effort: anthropic.BetaOutputConfigEffortMedium,
			Format: anthropic.BetaJSONOutputFormatParam{Schema: schema},
		},
		Messages: []anthropic.BetaMessageParam{anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(user))},
	})
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}
	g.log.Info("recipe generation", "model", resp.Model, "input_tokens", resp.Usage.InputTokens, "cache_read", resp.Usage.CacheReadInputTokens, "output_tokens", resp.Usage.OutputTokens)
	if resp.StopReason == anthropic.BetaStopReasonRefusal {
		return nil, errRefused
	}
	var text string
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			text += t.Text
		}
	}
	var out struct {
		Recipes []generatedCard `json:"recipes"`
	}
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return nil, fmt.Errorf("generate: parse answer: %w", err)
	}
	g.log.Debug("raw generation", "stop_reason", resp.StopReason, "text", text)
	recipes := []Recipe{}
	for _, c := range out.Recipes {
		r, err := buildRecipe(c, g.products)
		if err != nil {
			g.log.Warn("generated recipe dropped", "error", err)
			continue
		}
		recipes = append(recipes, r)
	}
	return recipes, nil
}

var monthsFR = []string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"}

func textOf(content []anthropic.ContentBlockUnion) string {
	var b strings.Builder
	for _, block := range content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}
