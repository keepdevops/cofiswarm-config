// Package build ports the config build tooling (build_swarm_config.py +
// migrate_swarm_config.py): assemble swarm-config.json from per-agent files, and
// split a monolithic swarm-config.json back into per-agent files. Output is
// byte-identical to Python's json.dumps(indent=2, sort_keys=True)+"\n".
package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// agentInternalKeys are stripped from per-agent files when assembling the monolith.
var agentInternalKeys = map[string]bool{"agent_id": true}

// requiredSplitFields mirrors migrate_swarm_config.REQUIRED_FIELDS.
var requiredSplitFields = []string{"name", "model", "system_prompt", "context", "max_tokens"}

func init() {
	if os.Getenv("MATRIX_MODEL_DIR") == "" {
		if runtime.GOOS == "darwin" {
			_ = os.Setenv("MATRIX_MODEL_DIR", "/Users/Shared/llama/models")
		} else {
			_ = os.Setenv("MATRIX_MODEL_DIR", "")
		}
	}
}

// encodePy serializes v exactly like Python json.dumps(v, indent=2, sort_keys=True)+"\n":
// 2-space indent, sorted map keys, no HTML (<>&) escaping, ensure_ascii (non-ASCII -> \uXXXX),
// one trailing newline.
func encodePy(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return asciiEscape(buf.Bytes()), nil
}

// asciiEscape replicates Python json.dumps' ensure_ascii=True: every rune >= 0x80 becomes
// \uXXXX (surrogate pair for > U+FFFF), lowercase hex. Non-ASCII only appears inside string
// literals in the encoder's output, so scanning the whole buffer is safe.
func asciiEscape(b []byte) []byte {
	var out bytes.Buffer
	for _, r := range string(b) {
		switch {
		case r < 0x80:
			out.WriteByte(byte(r))
		case r <= 0xFFFF:
			fmt.Fprintf(&out, `\u%04x`, r)
		default:
			r -= 0x10000
			fmt.Fprintf(&out, `\u%04x\u%04x`, 0xD800+(r>>10), 0xDC00+(r&0x3FF))
		}
	}
	return out.Bytes()
}

func loadJSON(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	// UseNumber preserves each number's source literal (e.g. 1.0 stays 1.0, 8086 stays 8086),
	// matching Python's int/float distinction on re-encode.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	return m, nil
}

func expandUser(p string) string {
	if p == "~" {
		h, _ := os.UserHomeDir()
		return h
	}
	if strings.HasPrefix(p, "~/") {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, p[2:])
	}
	return p
}

// expandModel mirrors os.path.expandvars(os.path.expanduser(model)) + the unresolved-$ guard.
func expandModel(model string) (string, error) {
	s := expandUser(model)
	s = os.Expand(s, func(k string) string {
		if v, ok := os.LookupEnv(k); ok {
			return v
		}
		return "${" + k + "}" // leave undefined vars (like Python expandvars), so the guard fires
	})
	if strings.Contains(s, "$") {
		return "", fmt.Errorf("unresolved env var in model path: %s", model)
	}
	return s, nil
}

// loadAgents reads config/agents/*.json (sorted), strips internal keys, expands model
// paths, and returns the agents sorted by name.
func loadAgents(agentsDir string) ([]map[string]any, error) {
	fi, err := os.Stat(agentsDir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("agents directory missing: %s", agentsDir)
	}
	paths, _ := filepath.Glob(filepath.Join(agentsDir, "*.json"))
	sort.Strings(paths)
	var agents []map[string]any
	for _, p := range paths {
		data, err := loadJSON(p)
		if err != nil {
			return nil, err
		}
		if _, ok := data["name"]; !ok {
			return nil, fmt.Errorf("%s missing 'name' field", p)
		}
		cleaned := map[string]any{}
		for k, v := range data {
			if !agentInternalKeys[k] {
				cleaned[k] = v
			}
		}
		if m, ok := cleaned["model"].(string); ok {
			exp, err := expandModel(m)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", p, err)
			}
			cleaned["model"] = exp
		}
		agents = append(agents, cleaned)
	}
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agent files under %s", agentsDir)
	}
	sort.SliceStable(agents, func(i, j int) bool {
		return agents[i]["name"].(string) < agents[j]["name"].(string)
	})
	return agents, nil
}

// Build assembles swarm-config.json from config/coordinator.json + config/agents/,
// writing it (and a public/ copy if that dir exists). Returns the agent count.
func Build(root string) (int, error) {
	coord, err := loadJSON(filepath.Join(root, "config", "coordinator.json"))
	if err != nil {
		return 0, err
	}
	agents, err := loadAgents(filepath.Join(root, "config", "agents"))
	if err != nil {
		return 0, err
	}
	out := map[string]any{
		"agents":      agents,
		"coordinator": orEmpty(coord["coordinator"]),
		"ui":          orEmpty(coord["ui"]),
	}
	if rag, ok := coord["rag"]; ok {
		out["rag"] = rag
	}
	payload, err := encodePy(out)
	if err != nil {
		return 0, err
	}
	outFile := filepath.Join(root, "swarm-config.json")
	if err := os.WriteFile(outFile, payload, 0o644); err != nil {
		return 0, fmt.Errorf("write %s: %w", outFile, err)
	}
	publicDir := filepath.Join(root, "public")
	if fi, err := os.Stat(publicDir); err == nil && fi.IsDir() {
		_ = os.WriteFile(filepath.Join(publicDir, "swarm-config.json"), payload, 0o644)
	}
	return len(agents), nil
}

func orEmpty(v any) any {
	if v == nil {
		return map[string]any{}
	}
	return v
}

func slugify(name string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-")
}

// Split splits a monolithic swarm-config.json into per-agent files (ports
// migrate_swarm_config.split_agents). Returns the number written.
func Split(source, outDir string, dryRun bool) (int, error) {
	data, err := loadJSON(source)
	if err != nil {
		return 0, err
	}
	rawAgents, ok := data["agents"].([]any)
	if !ok || len(rawAgents) == 0 {
		return 0, fmt.Errorf("no 'agents' array in %s", source)
	}
	if !dryRun {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return 0, err
		}
	}
	seen := map[string]bool{}
	written := 0
	for _, ra := range rawAgents {
		agent, ok := ra.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("agent entry is not an object")
		}
		for _, f := range requiredSplitFields {
			if _, ok := agent[f]; !ok {
				return 0, fmt.Errorf("agent %v missing field %q", agent["name"], f)
			}
		}
		slug := slugify(agent["name"].(string))
		if seen[slug] {
			return 0, fmt.Errorf("duplicate agent name after slugify: %s", slug)
		}
		seen[slug] = true
		payload := map[string]any{"agent_id": slug}
		for k, v := range agent {
			payload[k] = v
		}
		text, err := encodePy(payload)
		if err != nil {
			return 0, err
		}
		if !dryRun {
			if err := os.WriteFile(filepath.Join(outDir, slug+".json"), text, 0o644); err != nil {
				return 0, err
			}
		}
		written++
	}
	return written, nil
}
