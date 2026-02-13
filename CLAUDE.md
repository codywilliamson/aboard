# aboard

agent-powered kanban board TUI. go + bubbletea + trello api.

## quick reference

```bash
go build ./cmd/aboard        # build
go vet ./...                 # lint
go test ./...                # test
go run ./cmd/aboard          # run (needs .env with trello creds)
```

binary name is `aboard`. entry point is `cmd/aboard/main.go`.

## project structure

```
cmd/aboard/main.go           entry point — loads config, creates client/runner/model, runs tea program
internal/
  agent/runner.go            executes agent CLI commands (codex/claude), builds prompts with board context + action schema
  config/config.go           loads env vars + .env file, parses JSON command arrays
  trello/client.go           trello REST api — read (boards, lists, cards) + write (move, rename, comment, archive, create)
  ui/
    model.go                 root bubbletea model — state, Update routing, View composition, focus management (~800 lines, main file)
    messages.go              tea.Msg types (boardsLoaded, boardDataLoaded, agentResponse, cardMutated, listMutated) + tea.Cmd constructors
    kanban.go                KanbanModel — column rendering, per-list card cursors, horizontal scroll, ActiveList/SelectedCard helpers
    drawer.go                DrawerModel — card detail panel + scrollable timeline viewport (no input, prompt is in prompt bar)
    prompt.go                PromptBar — multi-mode input (agent/move/rename/comment/newcard/newlist/archive confirm)
    actions.go               parses <action>{json}</action> blocks from agent output, converts to mutation commands
    boards.go                board selector view + key handler
    help.go                  help overlay modal
    styles.go                lipgloss styles, badge helper, color palette
    util.go                  ellipsis, shortID, wrapForPane, clipToLineCount
docs/
  spec.md                    product overview + design principles
  architecture.md            codebase structure, data flow, focus system, mutation cycle
  roadmap.md                 versioned feature plan v0.2–v0.8
  backlog.md                 unscheduled future ideas
```

## architecture patterns

### bubbletea elm architecture
the app follows bubbletea's model-update-view pattern. `Model` is the root, sub-models (`KanbanModel`, `DrawerModel`, `PromptBar`) are embedded structs (not tea.Model implementations — they don't have their own Update). the root model routes all messages and keys.

### focus system
three focus areas: `focusKanban`, `focusDrawer`, `focusPrompt`. key handling routes through `handleKey` → `updateKanbanKeys` / `updateDrawerKeys` / `updatePromptKeys` based on current focus. esc cascades: prompt → drawer → kanban.

### mutation cycle
user keyboard action or agent `<action>` block → mutation command (`tea.Cmd`) → trello api call → `cardMutatedMsg`/`listMutatedMsg` → auto-refresh board via `loadBoardDataCmd`. errors surface in status bar + timeline.

### prompt bar modes
the prompt bar is a single component that switches between modes: `promptAgent`, `promptMove` (list picker), `promptRename`, `promptComment`, `promptNewCard`, `promptNewList`, `promptConfirmArchiveCard`, `promptConfirmArchiveList`. each mode has its own key handling in `updatePromptKeys`.

### agent action protocol
agents include `<action>{"type":"move_card",...}</action>` blocks. `parseActions()` in actions.go strips them from display text and returns structured actions. `executeActions()` converts to mutation commands via `tea.Batch`.

### trello client
form-encoded PUT/POST with key/token as form fields. `getJSON` for reads, `putForm`/`postForm` for writes. all methods take `context.Context` for timeout control.

## conventions

- go 1.23 — use builtins `max`/`min` (no custom helpers needed)
- lipgloss v1.0.0 — no `.Copy()` on styles, styles are value types
- bubbletea — value receiver on `Update` and `View`, pointer receiver on mutation helpers
- comments: lowercase, informal, minimal
- DRY/KISS first — minimal abstractions, no premature generalization
- interfaces defined where consumed (`trelloClient` in model.go, not in trello package)

## commits

**conventional commits required.** all commit messages must follow `<type>: <description>` format.

types: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `chore`, `style`

- `feat:` — new feature (bumps minor version)
- `fix:` — bug fix (bumps patch version)
- `feat!:` or `fix!:` or `BREAKING CHANGE:` in body — bumps major version
- all others — no version bump, included in changelog under their section

examples:
```
feat: add card label display in kanban columns
fix: correct column width calculation on narrow terminals
docs: update roadmap with v0.3 provider abstraction
refactor: extract board provider interface
chore: update bubbletea to v1.4.0
```

multi-line body for context when needed:
```
feat: add jira provider

implement BoardProvider interface for jira cloud rest api v3.
maps jira boards to kanban columns, issues to cards,
and transitions to move operations.
```

**do not** use scopes (no `feat(ui):`) — keep it flat and simple.

versioning is automated via release-please:
- push conventional commits to main
- release-please opens/updates a release PR with changelog + version bump
- merging the PR creates a git tag + github release
- goreleaser builds cross-platform binaries and attaches them

## key interfaces

```go
// in ui/model.go — what the UI needs from the trello client
type trelloClient interface {
    CanAuth() bool
    Boards(context.Context) ([]trello.Board, error)
    CardsForBoard(context.Context, string) ([]trello.Card, error)
    ListsForBoard(context.Context, string) ([]trello.List, error)
    MoveCard(context.Context, string, string) error
    UpdateCard(context.Context, string, string, string) error
    AddComment(context.Context, string, string) error
    ArchiveCard(context.Context, string) error
    CreateCard(context.Context, string, string) (*trello.Card, error)
    CreateList(context.Context, string, string) (*trello.List, error)
    ArchiveList(context.Context, string) error
}

// in ui/model.go — what the UI needs from the agent runner
type agentRunner interface {
    Ask(context.Context, agent.AgentName, string, string) (string, error)
}
```

## common tasks

### adding a new card operation
1. add trello api method in `trello/client.go` (putForm/postForm)
2. add method to `trelloClient` interface in `model.go`
3. add `tea.Cmd` constructor in `messages.go`
4. add prompt mode constant in `prompt.go` if it needs user input
5. add key handler in `updateKanbanKeys` (and `updateDrawerKeys` if applicable)
6. add starter method `startXxx()` in `model.go`
7. handle submission in `submitPrompt()` or dedicated submit method
8. mutation response is already handled by `cardMutatedMsg`/`listMutatedMsg`

### adding a new agent action type
1. add case in `executeActions()` switch in `actions.go`
2. add schema example in `buildPrompt()` in `agent/runner.go`
3. uses existing mutation command constructors from `messages.go`

### adding styles
all lipgloss styles live in `styles.go`. use existing color constants. the header uses plain styled text (not badge backgrounds) to avoid ANSI width measurement issues in `JoinHorizontal`.

## gotchas

- **lipgloss badge width**: background-colored badges break width calculation in `lipgloss.JoinHorizontal`. the header right side uses plain `statusStyle`/`subtleStyle` text instead. if you need a badge in a horizontally-joined layout, test it.
- **textinput prompt**: bubbletea's `textinput.New()` defaults to `Prompt: "> "`. the prompt bar sets `Prompt: " "` to avoid redundancy with the mode badge.
- **gitignore binary**: `/aboard` (with leading slash) to avoid matching `cmd/aboard/` directory.
- **value receivers**: bubbletea requires `Update(msg) (tea.Model, tea.Cmd)` with value receiver. internal mutation helpers use pointer receivers. the root model methods that return `(tea.Model, tea.Cmd)` use value receivers.
- **no tests yet**: the project has no test files. `go test ./...` passes (no test files = pass). tests are needed.

## configuration

config is loaded from a `.env` file. search order (first found wins):
1. `--config <path>` / `-c <path>` flag (explicit)
2. `.env` in current working directory
3. `<user config dir>/aboard/.env` (`%AppData%\aboard\.env` on windows, `~/.config/aboard/.env` on linux/mac)
4. `.env` next to the executable

env vars always override file values.

| var | required | description |
|-----|----------|-------------|
| `TRELLO_API_KEY` | yes | trello api key |
| `TRELLO_API_TOKEN` | yes | trello api token |
| `TRELLO_BOARD_ID` | no | skip board picker, load directly |
| `TRELLO_TUI_CODEX_COMMAND` | no | JSON array, default `["codex"]` |
| `TRELLO_TUI_CLAUDE_COMMAND` | no | JSON array, default `["claude"]` |

command arrays support `{prompt}` and `{context}` placeholders. if no placeholder, prompt is piped to stdin.

## ci

- `.github/workflows/ci.yml` — build + vet + test on push/PR to main
- `.github/workflows/release.yml` — release-please (auto version + changelog) + goreleaser (cross-platform binaries) on push to main
- `.goreleaser.yaml` — cross-compile linux/windows amd64/arm64, tar.gz + zip, appends to release-please releases
- `release-please-config.json` + `.release-please-manifest.json` — release-please configuration

releases are fully automated: push conventional commits → release-please PR → merge → tag + release + binaries.
