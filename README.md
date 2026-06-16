# cofiswarm-config

Cofiswarm component: `config`.

- Layout: [REPO-STANDARD-LAYOUT](https://github.com/keepdevops/cofiswarmdev/blob/main/docs/REPO-STANDARD-LAYOUT.md)
- Migration: [MIGRATION-SPRINTS](https://github.com/keepdevops/cofiswarmdev/blob/main/docs/MIGRATION-SPRINTS.md)

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
