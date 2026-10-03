package main

import (
	"encoding/json"
	"testing"
)

func TestMergePatch(t *testing.T) {
	tests := []struct {
		name, target, patch, want string
	}{
		{"add key", `{"a":1}`, `{"b":2}`, `{"a":1,"b":2}`},
		{"replace scalar", `{"a":1}`, `{"a":2}`, `{"a":2}`},
		{"delete with null", `{"a":1,"b":2}`, `{"a":null}`, `{"b":2}`},
		{"nested merge", `{"bought":{"x":true,"y":true}}`, `{"bought":{"z":true,"x":null}}`, `{"bought":{"y":true,"z":true}}`},
		{"array replaced", `{"deck":[1,2,3]}`, `{"deck":[4]}`, `{"deck":[4]}`},
		{"object over scalar", `{"a":1}`, `{"a":{"b":1}}`, `{"a":{"b":1}}`},
		{"null target", `null`, `{"a":{"b":null,"c":1}}`, `{"a":{"c":1}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target, patch, want any
			mustUnmarshal(t, tt.target, &target)
			mustUnmarshal(t, tt.patch, &patch)
			mustUnmarshal(t, tt.want, &want)

			got, _ := json.Marshal(mergePatch(target, patch))
			exp, _ := json.Marshal(want)
			if string(got) != string(exp) {
				t.Errorf("got %s, want %s", got, exp)
			}
		})
	}
}

func mustUnmarshal(t *testing.T, s string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(s), v); err != nil {
		t.Fatalf("unmarshal %s: %v", s, err)
	}
}
