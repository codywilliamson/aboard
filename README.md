# aboard

[![CI](https://github.com/codywilliamson/aboard/actions/workflows/ci.yml/badge.svg)](https://github.com/codywilliamson/aboard/actions/workflows/ci.yml)
[![Release](https://github.com/codywilliamson/aboard/actions/workflows/release.yml/badge.svg)](https://github.com/codywilliamson/aboard/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/codywilliamson/aboard)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Agent-powered kanban board TUI. Navigate boards, move cards, rename, comment, archive, and create — all from the terminal. Pipe prompts to local AI agents (`codex`, `claude`) that can read your board and execute actions.

```
┌─────────────────────────────────────────────┐
│ board name          agent:codex   status    │
├────────┬────────┬────────┬──────────────────┤
│ List 1 │ List 2 │ List 3 │ List 4           │
│>card a │ card d │ card g │                  │
│ card b │ card e │ card h │                  │
├────────┴────────┴────────┴──────────────────┤
│ [agent] ask the agent... █                  │
└─────────────────────────────────────────────┘
```

## Install

Download a binary from [Releases](https://github.com/codywilliamson/aboard/releases), or build from source:

```bash
go install github.com/codywilliamson/aboard/cmd/aboard@latest
```

## Setup

Create a `.env` file or export these variables:

```bash
export TRELLO_API_KEY=your_key
export TRELLO_API_TOKEN=your_token
# optional — skip the board picker
export TRELLO_BOARD_ID=abc123
```

Get your API key and token from [trello.com/power-ups/admin](https://trello.com/power-ups/admin).

### Agent commands (optional)

Configure which CLI agents to use:

```bash
export TRELLO_TUI_CODEX_COMMAND='["codex","exec","--skip-git-repo-check","{prompt}"]'
export TRELLO_TUI_CLAUDE_COMMAND='["claude","-p","{prompt}"]'
```

Placeholders: `{prompt}` (full prompt with board context), `{context}` (board context only). If no placeholder, prompt is piped to stdin.

## Usage

```bash
aboard
```

## Keyboard shortcuts

| Context | Key | Action |
|---------|-----|--------|
| Kanban | `h`/`l` | Move between columns |
| Kanban | `j`/`k` | Move within column |
| Kanban | `enter` | Open card in drawer |
| Kanban | `m` | Move card (list picker) |
| Kanban | `e` | Rename card |
| Kanban | `c` | Comment on card |
| Kanban | `n` | New card in current list |
| Kanban | `N` | New list on board |
| Kanban | `x` | Archive card |
| Kanban | `X` | Archive list |
| Kanban | `/` or `tab` | Focus prompt bar |
| Kanban | `1`/`2`/`a` | Agent controls |
| Kanban | `r`/`b`/`q` | Refresh / boards / quit |
| Prompt | `enter` | Submit |
| Prompt | `esc` or `tab` | Cancel / return |
| Drawer | `j`/`k` | Scroll timeline |
| Drawer | `tab` | Focus kanban |
| Drawer | `/` | Focus prompt |
| Global | `ctrl+c` | Quit |
| Global | `?` | Help overlay |

## Agent actions

Agents can perform board mutations by including action blocks in their response:

```
<action>{"type":"move_card","card_id":"...","list_id":"..."}</action>
<action>{"type":"create_card","list_id":"...","name":"..."}</action>
```

Supported: `move_card`, `update_card`, `add_comment`, `archive_card`, `create_card`, `create_list`, `archive_list`.

## Build from source

```bash
go build ./cmd/aboard
```

## License

MIT
