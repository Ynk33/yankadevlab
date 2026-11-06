package main

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
)

const (
	portions          = 2
	bunchGrams        = 50
	defaultPieceGrams = 100
)

var spoonML = map[string]float64{"c. à s.": 15, "c. à c.": 5}

var meatAndFish = []string{"poisson", "fruits de mer", "volaille", "viande rouge", "porc"}

// Product is one entry of data/products.json, exported by import-pipeline/15_exportProducts.py.
type Product struct {
	NameEN    string             `json:"name_en"`
	Rayon     string             `json:"rayon"`
	Role      string             `json:"role"`
	Protein   string             `json:"protein"`
	Unit      string             `json:"unit"`
	Density   float64            `json:"density_g_per_ml"`
	PieceG    float64            `json:"piece_g"`
	Nutrients map[string]float64 `json:"nutrients"`
}

// Ingredient is a recipe ingredient in catalog units, for 2 people.
type Ingredient struct {
	Name string  `json:"name"`
	Qty  float64 `json:"qty"`
	Unit string  `json:"unit"`
}

func loadProducts() (map[string]Product, error) {
	raw, err := dataFS.ReadFile("data/products.json")
	if err != nil {
		return nil, err
	}
	products := map[string]Product{}
	if err := json.Unmarshal(raw, &products); err != nil {
		return nil, fmt.Errorf("parse products.json: %w", err)
	}
	return products, nil
}

// grams converts an amount in a catalog unit to grams with the product's density and piece weight.
func grams(p Product, qty float64, unit string) float64 {
	density := p.Density
	if density <= 0 {
		density = 1
	}
	switch unit {
	case "g":
		return qty
	case "ml":
		return qty * density
	case "pièce":
		if p.PieceG > 0 {
			return qty * p.PieceG
		}
		return qty * defaultPieceGrams
	case "botte":
		return qty * bunchGrams
	default:
		return qty * spoonML[unit] * density
	}
}

// nutrientsPerPortion sums the Ciqual nutrients (per 100 g) of the ingredients for one portion.
func nutrientsPerPortion(ingredients []Ingredient, products map[string]Product) map[string]float64 {
	totals := map[string]float64{}
	for _, in := range ingredients {
		p := products[in.Name]
		g := grams(p, in.Qty, in.Unit)
		for nutrient, per100 := range p.Nutrients {
			totals[nutrient] += per100 * g / 100 / portions
		}
	}
	return totals
}

// mainProtein returns the heaviest protein category of the recipe and its grams per portion, or "none".
func mainProtein(ingredients []Ingredient, products map[string]Product) (string, float64) {
	totals := map[string]float64{}
	for _, in := range ingredients {
		p := products[in.Name]
		if (p.Role == "protein" || p.Role == "legume" || p.Role == "cheese") && p.Protein != "none" {
			totals[p.Protein] += grams(p, in.Qty, in.Unit) / portions
		}
	}
	best, bestG := "none", 0.0
	for category, g := range totals {
		if g > bestG || (g == bestG && category < best) {
			best, bestG = category, g
		}
	}
	return best, bestG
}

type highlightRule struct {
	label, nutrient string
	threshold       float64
}

// Thresholds per portion, as in import-pipeline/utils/11_nutrition/highlights.py: micronutrients at 50% of the
// EU reference intake, proteins 30 g, fibres 8 g.
var highlightRules = []highlightRule{
	{"Vitamine C", "vitamine_c", 40}, {"Fer", "fer", 7}, {"Calcium", "calcium", 400}, {"Magnésium", "magnesium", 187.5},
	{"Potassium", "potassium", 1000}, {"Zinc", "zinc", 5}, {"Iode", "iode", 75}, {"Sélénium", "selenium", 27.5},
	{"Folates", "folates", 100}, {"Vitamine B12", "vitamine_b12", 1.25}, {"Vitamine D", "vitamine_d", 2.5},
	{"Vitamine K", "vitamine_k", 37.5}, {"Bêta-carotène", "beta_carotene", 2400},
	{"Protéines", "proteines", 30}, {"Fibres", "fibres", 8},
}

var highlightEN = map[string]string{
	"Oméga-3": "Omega-3", "Fer": "Iron", "Folates": "Folate", "Vitamine K": "Vitamin K", "Vitamine D": "Vitamin D",
	"Bêta-carotène": "Beta-carotene", "Calcium": "Calcium", "Iode": "Iodine", "Magnésium": "Magnesium",
	"Vitamine B12": "Vitamin B12", "Fibres": "Fibre", "Vitamine C": "Vitamin C", "Potassium": "Potassium",
	"Sélénium": "Selenium", "Protéines": "Protein", "Zinc": "Zinc",
}

// highlights picks up to 3 nutritional benefits of a portion, strongest first. Vitamin B12 only counts without meat
// or fish, where nearly any portion reaches it.
func highlights(n map[string]float64, hasMeatOrFish bool) []string {
	type ratio struct {
		label string
		value float64
	}
	var reached []ratio
	for _, rule := range highlightRules {
		if rule.label == "Vitamine B12" && hasMeatOrFish {
			continue
		}
		if v := n[rule.nutrient] / rule.threshold; v >= 1 {
			reached = append(reached, ratio{rule.label, v})
		}
	}
	if v := math.Max((n["epa"]+n["dha"])/0.25, n["ala"]/1.1); v >= 1 {
		reached = append(reached, ratio{"Oméga-3", v})
	}
	slices.SortStableFunc(reached, func(a, b ratio) int {
		switch {
		case a.value > b.value:
			return -1
		case a.value < b.value:
			return 1
		}
		return 0
	})
	out := []string{}
	for i := 0; i < len(reached) && i < 3; i++ {
		out = append(out, reached[i].label)
	}
	return out
}
