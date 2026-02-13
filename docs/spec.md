# aboard spec

a cross-platform terminal kanban board with integrated AI agents. navigate boards, manage cards, and delegate work to local coding agents — all without leaving the terminal.

currently backed by trello. architecture is being shaped toward provider-agnostic backends (jira, github projects, linear, local/offline).

platforms: linux (amd64/arm64), windows (amd64/arm64)

## docs

- [architecture](architecture.md) — current codebase structure + data flow
- [roadmap](roadmap.md) — versioned feature plan (v0.2–v0.8)
- [backlog](backlog.md) — unscheduled future ideas

## design principles

1. **keyboard-first** — every action reachable without a mouse
2. **agent-native** — agents are first-class citizens, not bolted on
3. **provider-agnostic** — trello today, anything tomorrow
4. **single binary** — no runtime dependencies, no install scripts
5. **fast** — sub-second startup, no loading spinners for local ops
6. **minimal config** — works with just api credentials, everything else optional
