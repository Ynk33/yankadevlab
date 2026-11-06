package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed data/products.json web/src/data/catalog.json web/src/data/imported.json
var dataFS embed.FS

// RecipeText is the translated part of a recipe card.
type RecipeText struct {
	Title      string   `json:"title"`
	Highlights []string `json:"highlights"`
	Steps      []string `json:"steps"`
}

// RecipeIngredient is an ingredient as the front end stores it (CustomRecipeJSON in web/src/lib/catalog.ts).
type RecipeIngredient struct {
	Ingredient
	Rayon  string `json:"rayon"`
	NameEN string `json:"name_en"`
}

// Recipe is a generated recipe, shaped like CustomRecipeJSON in web/src/lib/catalog.ts so the front end registers
// it with registerCustom.
type Recipe struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Meal        string             `json:"meal"`
	Protein     string             `json:"protein"`
	Minutes     int                `json:"minutes"`
	Kcal        int                `json:"kcal"`
	Proteines   int                `json:"proteines"`
	Fibres      int                `json:"fibres"`
	Highlights  []string           `json:"highlights"`
	Ingredients []RecipeIngredient `json:"ingredients"`
	Steps       []string           `json:"steps"`
	World       bool               `json:"world"`
	Batch       bool               `json:"batch"`
	Base        int                `json:"base"`
	EN          RecipeText         `json:"en"`
}

// catalogEntry is the compact view of a catalog recipe given to the matching model.
type catalogEntry struct {
	ID, Title, Meal, Protein string
	Minutes                  int
	World, Batch             bool
	Ingredients              []string
}

func (e catalogEntry) line() string {
	tags := []string{e.Meal}
	if e.World {
		tags = append(tags, "world")
	}
	if e.Batch {
		tags = append(tags, "batch")
	}
	return fmt.Sprintf("%s | %s | %s | %d min | %s | %s", e.ID, e.Title, e.Protein, e.Minutes, strings.Join(tags, ","), strings.Join(e.Ingredients, ", "))
}

// staticCatalog reads the hand-written catalog.json and the pipeline's imported.json (embedded) as compact entries.
func staticCatalog() ([]catalogEntry, error) {
	var entries []catalogEntry
	seen := map[string]bool{}
	for _, file := range []string{"web/src/data/catalog.json", "web/src/data/imported.json"} {
		raw, err := dataFS.ReadFile(file)
		if err != nil {
			return nil, err
		}
		var doc struct {
			Recipes [][]json.RawMessage `json:"recipes"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
		for _, t := range doc.Recipes {
			e, err := tupleEntry(t)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file, err)
			}
			if !seen[e.ID] {
				seen[e.ID] = true
				entries = append(entries, e)
			}
		}
	}
	return entries, nil
}

// tupleEntry decodes a RawRecipe tuple (web/src/lib/catalog.ts): id, title, meal, protein, minutes, kcal, proteines,
// fibres, highlights, ingredients, steps, world, batch.
func tupleEntry(t []json.RawMessage) (catalogEntry, error) {
	var e catalogEntry
	if len(t) != 13 {
		return e, fmt.Errorf("recipe tuple has %d fields", len(t))
	}
	var ingredients [][]json.RawMessage
	var world, batch int
	for i, dst := range map[int]any{0: &e.ID, 1: &e.Title, 2: &e.Meal, 3: &e.Protein, 4: &e.Minutes, 9: &ingredients, 11: &world, 12: &batch} {
		if err := json.Unmarshal(t[i], dst); err != nil {
			return e, fmt.Errorf("recipe field %d: %w", i, err)
		}
	}
	for _, in := range ingredients {
		var name string
		if len(in) > 0 && json.Unmarshal(in[0], &name) == nil {
			e.Ingredients = append(e.Ingredients, name)
		}
	}
	e.World, e.Batch = world == 1, batch == 1
	return e, nil
}

func recipeEntry(r Recipe) catalogEntry {
	names := make([]string, len(r.Ingredients))
	for i, in := range r.Ingredients {
		names[i] = in.Name
	}
	return catalogEntry{ID: r.ID, Title: r.Title, Meal: r.Meal, Protein: r.Protein, Minutes: r.Minutes, World: r.World, Batch: r.Batch, Ingredients: names}
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slug mirrors slug() in web/src/lib/utils.ts: lowercase, accents stripped, non-alphanumerics to "-", 80 chars max.
func slug(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(s)) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(nonSlug.ReplaceAllString(b.String(), "-"), "-")
	if len(out) > 80 {
		out = strings.TrimRight(out[:80], "-")
	}
	return out
}

func newRecipeID(title string) string {
	suffix := make([]byte, 3)
	rand.Read(suffix)
	base := slug(title)
	if len(base) > 60 {
		base = strings.TrimRight(base[:60], "-")
	}
	return "gen-" + base + "-" + hex.EncodeToString(suffix)
}

// roundQty rounds like import-pipeline/utils/08_scaleServings/round_quantity.py: 5 g/ml steps under 100, 10 above;
// halves for pieces, bunches and spoons.
func roundQty(qty float64, unit string) float64 {
	if unit == "g" || unit == "ml" {
		step := 5.0
		if qty >= 100 {
			step = 10
		}
		return math.Max(step, math.Round(qty/step)*step)
	}
	return math.Max(0.5, math.Round(qty*2)/2)
}

// generatedCard is what the generation model writes for one recipe.
type generatedCard struct {
	Title       string `json:"title"`
	TitleEN     string `json:"title_en"`
	Meal        string `json:"meal"`
	Minutes     int    `json:"minutes"`
	World       bool   `json:"world"`
	Batch       bool   `json:"batch"`
	Ingredients []struct {
		Name string  `json:"name"`
		Qty  float64 `json:"qty"`
	} `json:"ingredients"`
	Steps   []string `json:"steps"`
	StepsEN []string `json:"steps_en"`
}

// buildRecipe turns a generated card into a catalog recipe: ingredients restricted to known products in their
// catalog unit, nutrition and protein category computed from the product table. It fails when the card is unusable.
func buildRecipe(card generatedCard, products map[string]Product) (Recipe, error) {
	if strings.TrimSpace(card.Title) == "" || len(card.Steps) < 3 || len(card.Steps) != len(card.StepsEN) || card.Minutes <= 0 {
		return Recipe{}, fmt.Errorf("incomplete card %q", card.Title)
	}
	var ingredients []Ingredient
	var withMeta []RecipeIngredient
	for _, in := range card.Ingredients {
		p, ok := products[in.Name]
		if !ok || in.Qty <= 0 || slices.ContainsFunc(ingredients, func(x Ingredient) bool { return x.Name == in.Name }) {
			continue
		}
		ing := Ingredient{Name: in.Name, Qty: roundQty(in.Qty, p.Unit), Unit: p.Unit}
		ingredients = append(ingredients, ing)
		withMeta = append(withMeta, RecipeIngredient{Ingredient: ing, Rayon: p.Rayon, NameEN: p.NameEN})
	}
	if len(ingredients) < 3 {
		return Recipe{}, fmt.Errorf("%q has fewer than 3 known ingredients", card.Title)
	}
	protein, _ := mainProtein(ingredients, products)
	if protein == "none" {
		return Recipe{}, fmt.Errorf("%q has no main protein", card.Title)
	}
	n := nutrientsPerPortion(ingredients, products)
	hasMeatOrFish := slices.ContainsFunc(ingredients, func(in Ingredient) bool { return slices.Contains(meatAndFish, products[in.Name].Protein) })
	labels := highlights(n, hasMeatOrFish)
	labelsEN := make([]string, len(labels))
	for i, l := range labels {
		labelsEN[i] = highlightEN[l]
	}
	meal := card.Meal
	if meal != "midi" && meal != "soir" {
		meal = "both"
	}
	return Recipe{
		ID: newRecipeID(card.Title), Title: card.Title, Meal: meal, Protein: protein, Minutes: card.Minutes,
		Kcal: int(math.Round(n["kcal"])), Proteines: int(math.Round(n["proteines"])), Fibres: int(math.Round(n["fibres"])),
		Highlights: labels, Ingredients: withMeta, Steps: card.Steps, World: card.World, Batch: card.Batch, Base: portions,
		EN: RecipeText{Title: card.TitleEN, Highlights: labelsEN, Steps: card.StepsEN},
	}, nil
}
