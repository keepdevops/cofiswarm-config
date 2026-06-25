package build

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	r, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Build must reproduce the committed swarm-config.json byte-for-byte (i.e. it matches
// the Python build_swarm_config.py that generated it).
func TestBuildByteIdenticalToCommitted(t *testing.T) {
	repo := repoRoot(t)
	tmp := t.TempDir()
	if err := os.CopyFS(filepath.Join(tmp, "config"), os.DirFS(filepath.Join(repo, "config"))); err != nil {
		t.Fatalf("copy config: %v", err)
	}
	if _, err := Build(tmp); err != nil {
		t.Fatalf("build: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(tmp, "swarm-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(repo, "swarm-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Go build differs from committed swarm-config.json (got %d bytes, want %d) — "+
			"either the build is non-faithful or swarm-config.json is stale vs config/agents/",
			len(got), len(want))
	}
}

// Split of the committed swarm-config.json must reproduce the committed config/agents/*.json.
func TestSplitByteIdenticalToCommitted(t *testing.T) {
	repo := repoRoot(t)
	tmp := t.TempDir()
	n, err := Split(filepath.Join(repo, "swarm-config.json"), tmp, false)
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if n == 0 {
		t.Fatal("split wrote 0 files")
	}
	entries, _ := filepath.Glob(filepath.Join(tmp, "*.json"))
	for _, got := range entries {
		name := filepath.Base(got)
		want := filepath.Join(repo, "config", "agents", name)
		g, _ := os.ReadFile(got)
		w, err := os.ReadFile(want)
		if err != nil {
			t.Errorf("split produced %s with no committed counterpart", name)
			continue
		}
		if !bytes.Equal(g, w) {
			t.Errorf("split %s differs from committed config/agents/%s", name, name)
		}
	}
}
