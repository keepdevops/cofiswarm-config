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
