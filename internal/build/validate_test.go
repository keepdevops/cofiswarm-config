package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEffectiveEngine(t *testing.T) {
	cases := []struct {
		a    map[string]any
		want string
	}{
		{map[string]any{}, "llama"},                                    // default
		{map[string]any{"engine": "mlx"}, "mlx"},                       // engine
		{map[string]any{"engine": "llama", "backend": "vllm"}, "vllm"}, // backend overrides
	}
	for _, c := range cases {
		if got := effectiveEngine(c.a); got != c.want {
			t.Errorf("effectiveEngine(%v) = %q, want %q", c.a, got, c.want)
		}
	}
}

func TestValidateAgent(t *testing.T) {
	ok := []map[string]any{
		{"engine": "llama", "model": "/Users/Shared/llama/models/x.gguf"},
		{"engine": "mlx", "model": "/Users/Shared/llama/models/MLX/x"},
		{"engine": "vllm", "model": "Qwen2.5-7B-Instruct-4bit"},
		{}, // default llama, no model required
	}
	for _, a := range ok {
		if err := validateAgent("a", a); err != nil {
			t.Errorf("validateAgent(%v) unexpected error: %v", a, err)
		}
	}

	bad := map[string]struct {
		a    map[string]any
		want string // substring the error should mention
	}{
		"unknown engine":   {map[string]any{"engine": "vlmm"}, "unknown engine"},
		"vllm no model":    {map[string]any{"engine": "vllm"}, "requires a non-empty"},
		"vllm empty model": {map[string]any{"engine": "vllm", "model": "  "}, "requires a non-empty"},
		"vllm path model":  {map[string]any{"engine": "vllm", "model": "/Users/Shared/x.gguf"}, "not a filesystem path"},
		"vllm via backend": {map[string]any{"backend": "vllm"}, "requires a non-empty"},
	}
	for name, c := range bad {
		err := validateAgent("a", c.a)
		if err == nil {
			t.Errorf("%s: expected an error, got nil", name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %q does not mention %q", name, err.Error(), c.want)
		}
	}
}

// Validation must be wired into Build, not just unit-testable in isolation.
func TestBuildRejectsBadVLLMAgent(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("config/coordinator.json", `{}`)
	write("config/agents/bad.json",
		`{"name":"bad","engine":"vllm","model":"/Users/Shared/x.gguf","system_prompt":"x","max_tokens":16}`)

	if _, err := Build(root); err == nil {
		t.Fatal("Build should reject a vllm agent whose model is a filesystem path")
	} else if !strings.Contains(err.Error(), "not a filesystem path") {
		t.Errorf("unexpected error: %v", err)
	}
}
