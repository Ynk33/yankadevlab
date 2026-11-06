package main

import (
	"slices"
	"strings"
	"testing"
)

var testProducts = map[string]Product{
	"saumon":        {Rayon: "Boucherie & poisson", Role: "protein", Protein: "poisson", Unit: "g", Density: 1, Nutrients: map[string]float64{"kcal": 200, "proteines": 20, "vitamine_b12": 3}},
	"poireau":       {Rayon: "Fruits & légumes", Role: "vegetable", Protein: "none", Unit: "pièce", PieceG: 150, Nutrients: map[string]float64{"kcal": 30, "fibres": 2}},
	"huile d'olive": {Rayon: "Épicerie", Role: "fat", Protein: "none", Unit: "c. à s.", Density: 0.92, Nutrients: map[string]float64{"kcal": 900}},
	"œuf":           {Rayon: "Crèmerie & œufs", Role: "protein", Protein: "œufs", Unit: "pièce", PieceG: 55, Nutrients: map[string]float64{"vitamine_b12": 2}},
}

func card(names ...string) generatedCard {
	c := generatedCard{Title: "Saumon aux poireaux", TitleEN: "Salmon with leeks", Meal: "soir", Minutes: 25,
		Steps: []string{"a", "b", "c"}, StepsEN: []string{"a", "b", "c"}}
	for _, n := range names {
		c.Ingredients = append(c.Ingredients, struct {
			Name string  `json:"name"`
			Qty  float64 `json:"qty"`
		}{n, 2.2})
	}
	return c
}

func TestGramsUsesCatalogUnitConversions(t *testing.T) {
	cases := []struct {
		unit string
		want float64
	}{{"g", 2}, {"pièce", 300}, {"botte", 100}, {"c. à s.", 27.6}, {"c. à c.", 9.2}}
	p := Product{PieceG: 150, Density: 0.92}
	for _, c := range cases {
		if got := grams(p, 2, c.unit); got < c.want-0.01 || got > c.want+0.01 {
			t.Errorf("grams(2 %s) = %v, want %v", c.unit, got, c.want)
		}
	}
}

func TestHighlightsSkipsB12WithMeatOrFish(t *testing.T) {
	n := map[string]float64{"vitamine_b12": 5, "proteines": 40}
	if got := highlights(n, true); slices.Contains(got, "Vitamine B12") {
		t.Errorf("B12 kept with fish: %v", got)
	}
	if got := highlights(n, false); got[0] != "Vitamine B12" {
		t.Errorf("B12 should lead without meat or fish: %v", got)
	}
}

func TestBuildRecipeForcesUnitsAndComputesNutrition(t *testing.T) {
	r, err := buildRecipe(card("saumon", "poireau", "huile d'olive", "inconnu"), testProducts)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Ingredients) != 3 {
		t.Fatalf("unknown ingredient kept: %+v", r.Ingredients)
	}
	if r.Ingredients[0].Unit != "g" || r.Ingredients[0].Qty != 5 || r.Ingredients[1].Unit != "pièce" || r.Ingredients[1].Qty != 2 {
		t.Errorf("units or rounding not applied: %+v", r.Ingredients)
	}
	if r.Protein != "poisson" || r.Base != 2 || !strings.HasPrefix(r.ID, "gen-saumon-aux-poireaux-") || r.EN.Title != "Salmon with leeks" {
		t.Errorf("unexpected recipe: %+v", r)
	}
	if r.Kcal <= 0 || r.Ingredients[0].Rayon != "Boucherie & poisson" {
		t.Errorf("nutrition or aisle missing: kcal=%d %+v", r.Kcal, r.Ingredients[0])
	}
}

func TestBuildRecipeRejectsUnusableCards(t *testing.T) {
	if _, err := buildRecipe(card("poireau", "huile d'olive", "poireau"), testProducts); err == nil {
		t.Error("card without a main protein accepted")
	}
	c := card("saumon", "poireau", "huile d'olive")
	c.StepsEN = c.StepsEN[:2]
	if _, err := buildRecipe(c, testProducts); err == nil {
		t.Error("card with untranslated steps accepted")
	}
}

func TestSlugMatchesFrontEnd(t *testing.T) {
	if got := slug("Saumon rôti, lentilles vertes et épinards"); got != "saumon-roti-lentilles-vertes-et-epinards" {
		t.Errorf("slug = %q", got)
	}
}

func TestEmbeddedDataLoads(t *testing.T) {
	products, err := loadProducts()
	if err != nil || len(products) < 100 {
		t.Fatalf("products: %d, %v", len(products), err)
	}
	entries, err := staticCatalog()
	if err != nil || len(entries) < 178 {
		t.Fatalf("catalog: %d, %v", len(entries), err)
	}
	for _, e := range entries {
		if e.ID == "" || e.Title == "" || len(e.Ingredients) == 0 {
			t.Fatalf("incomplete entry: %+v", e)
		}
	}
}
