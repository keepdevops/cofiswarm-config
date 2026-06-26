# cofiswarm-config

Cofiswarm component: `config`. Source of truth for agent configs (`config/agents/*.json`)
and the assembled `swarm-config.json`.

- Layout: [REPO-STANDARD-LAYOUT](https://github.com/keepdevops/cofiswarm-docs/blob/main/REPO-STANDARD-LAYOUT.md)
- Migration: [MIGRATION-SPRINTS](https://github.com/keepdevops/cofiswarm-docs/blob/main/MIGRATION-SPRINTS.md)

## Build tool (Go)

`cmd/cofiswarm-config` (`make build` → `bin/cofiswarm-config`) — Go port of the former
`scripts/build_swarm_config.py` + `scripts/migrate_swarm_config.py`. Output is **byte-identical**
to the Python `json.dumps(indent=2, sort_keys=True)` (ensure_ascii + number-literal preserved).

```bash
cofiswarm-config build [--root DIR]   # config/agents/*.json + coordinator.json -> swarm-config.json
cofiswarm-config split [--source FILE] [--out-dir DIR] [--dry-run]  # swarm-config.json -> per-agent files
```

## Agent `engine` + `model`

Each `config/agents/<name>.json` declares an inference engine and a model. The
build validates engine/model coherence and **fails loudly** on a mismatch (a
typo'd engine or a vLLM agent without a served model id) rather than deferring
the error to inference time.

| `engine` | `model` means | reachability |
|----------|---------------|--------------|
| `llama` (default) | weights path (`.gguf`); `~`/`$VAR` are expanded | host-local |
| `mlx` | weights directory path; `~`/`$VAR` expanded | host-local (Apple Metal) |
| `vllm` | **served model id** (e.g. `Qwen2.5-7B-Instruct-4bit`), *not* a path | shared server (`:12434`) |

An explicit `"backend"` overrides `"engine"`. Known engines: `llama`, `mlx`,
`vllm`, `ollama`, `sglang`. The modes consume `model` only for `vllm`; llama/mlx
ignore it at the mode layer (the path is used to launch the server). See
[`templates/agent-vllm.example.json`](templates/agent-vllm.example.json) for a
vLLM agent template — copy it into `config/agents/` to deploy.

## FHS paths

| Path | Purpose |
|------|---------|
| `/etc/cofiswarm/config/` | config |
| `/var/lib/cofiswarm/config/` | state |
| `/var/log/cofiswarm/config/` | logs |

## Test

```bash
./test/scripts/assert-layout.sh config
```
