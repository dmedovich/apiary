package main

import (
	"bytes"
	"os"
	"testing"
)

// TestGolden asserts that the full generation pipeline reproduces the committed
// docs/*.yaml byte-for-byte, using the same flags as `make generate`. It is the
// regression guard for generated output: any change to the generator or the
// example sources must be accompanied by regenerated docs.
//
// To refresh the snapshots: `make generate` (or `UPDATE_GOLDEN=1 go test ./cmd/apiary`).
func TestGolden(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string
		opt      specOptions
		golden   string
	}{
		{
			name:     "sample",
			patterns: []string{"../../testdata/sample"},
			opt:      specOptions{title: "Sample API", version: "0.1.0"},
			golden:   "../../docs/sample.yaml",
		},
		{
			name:     "router",
			patterns: []string{"../../testdata/router"},
			opt:      specOptions{title: "Task Manager API", version: "1.0.0", security: []string{"bearer"}},
			golden:   "../../docs/tasks.yaml",
		},
		{
			// gin example lives in a nested module (its own go.mod with the real
			// gin dependency), so it is loaded from its own directory.
			name:     "gin",
			patterns: []string{"."},
			opt:      specOptions{title: "Task Manager API (gin)", version: "1.0.0", security: []string{"bearer"}, dir: "../../testdata/gin"},
			golden:   "../../docs/tasks_gin.yaml",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spec, n, err := buildSpec(c.patterns, c.opt)
			if err != nil {
				t.Fatalf("buildSpec: %v", err)
			}
			if n == 0 {
				t.Fatal("no operations found")
			}
			got, err := encodeSpec(spec, "yaml")
			if err != nil {
				t.Fatalf("encodeSpec: %v", err)
			}

			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(c.golden, got, 0o644); err != nil {
					t.Fatalf("update golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(c.golden)
			if err != nil {
				t.Fatalf("read golden %s: %v (run `make generate` to create it)", c.golden, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("generated spec differs from %s — run `make generate` to regenerate", c.golden)
			}
		})
	}
}
