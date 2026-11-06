//go:build live

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"
)

// TestGeneratorLive calls the real API (a few cents): go test -tags live -run TestGeneratorLive -v
// with ANTHROPIC_API_KEY set and ASSIETTE_REQUEST optionally overriding the request.
func TestGeneratorLive(t *testing.T) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}
	request := os.Getenv("ASSIETTE_REQUEST")
	if request == "" {
		request = "j'ai des poireaux et du saumon, envie de quelque chose de réconfortant, 30 minutes max"
	}
	g, err := NewGenerator(key, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	start := time.Now()
	ids, err := g.Match(ctx, request, nil, nil, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("matches (%s): %v", time.Since(start).Round(time.Millisecond), ids)
	start = time.Now()
	var avoid []string
	for _, e := range g.catalog {
		if slices.Contains(ids, e.ID) {
			avoid = append(avoid, e.Title)
		}
	}
	n := 3
	if v, err := strconv.Atoi(os.Getenv("ASSIETTE_N")); err == nil {
		n = v
	}
	ids, err = g.Match(ctx, request, nil, nil, n)
	if err != nil {
		t.Fatal(err)
	}
	recipes, err := g.Generate(ctx, request, avoid, n-len(ids), time.Now().Month())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("generated %d recipes in %s", len(recipes), time.Since(start).Round(time.Millisecond))
	for _, r := range recipes {
		out, _ := json.MarshalIndent(r, "", " ")
		t.Log(string(out))
	}
}
