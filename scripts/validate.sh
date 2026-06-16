#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
[[ -d "${ROOT}/config/agents" ]] || { echo "missing config/agents/"; exit 1; }
[[ -f "${ROOT}/config/coordinator.json" ]] || { echo "missing config/coordinator.json"; exit 1; }
python3 "${ROOT}/scripts/build_swarm_config.py" --root "${ROOT}"
echo "ok: config validates"
