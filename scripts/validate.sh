#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
[[ -d "${ROOT}/agents" ]] || { echo "missing agents/"; exit 1; }
[[ -f "${ROOT}/coordinator.json" ]] || { echo "missing coordinator.json"; exit 1; }
python3 "${ROOT}/scripts/build_swarm_config.py" -o /tmp/swarm-config.validate.json
echo "ok: config validates"
