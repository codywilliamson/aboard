# roadmap

## v0.2 — polish + ux

**card labels + due dates**
- fetch and display trello labels (color dots) on card rows
- show due dates in card detail and as subtle indicators in columns
- filter cards by label (toggle filter mode)

**search**
- `/` in agent mode types a prompt, but add `ctrl+f` for card search
- fuzzy match across all cards on the board, jump to result

**card description editing**
- `d` on a card opens prompt bar in description-edit mode
- prefill with current description, submit updates via api

**streaming agent output**
- currently waits for full agent response before displaying
- stream stdout line-by-line into the timeline viewport
- show a spinner/progress indicator while agent is running

**better timeline**
- syntax highlighting for code blocks in agent responses
- collapsible entries (fold old exchanges)
- copy timeline entry to clipboard

## v0.3 — provider abstraction

**board provider interface**
```go
type BoardProvider interface {
    Name() string
    Boards(ctx) ([]Board, error)
    Lists(ctx, boardID) ([]List, error)
    Cards(ctx, boardID) ([]Card, error)
    MoveCard(ctx, cardID, listID) error
    UpdateCard(ctx, cardID, fields) error
    AddComment(ctx, cardID, text) error
    ArchiveCard(ctx, cardID) error
    CreateCard(ctx, listID, name) (*Card, error)
    CreateList(ctx, boardID, name) (*List, error)
    ArchiveList(ctx, listID) error
}
```

**generic data models**
- `Board`, `List`, `Card` become provider-agnostic structs in `internal/board/`
- trello client implements the interface, maps trello-specific fields
- ui layer only talks to the interface

**provider selection**
- config: `ABOARD_PROVIDER=trello|jira|github|local`
- board selector shows provider name
- agent context includes provider-specific metadata

## v0.4 — jira provider

- jira cloud rest api v3 integration
- boards = jira boards, lists = columns/statuses, cards = issues
- map jira transitions to move operations
- jira-specific fields: assignee, priority, sprint
- auth via api token or oauth

## v0.5 — github projects provider

- github projects v2 via graphql api
- projects = boards, status field values = lists, items = cards
- leverage existing `gh` cli auth
- support draft issues + conversion to real issues

## v0.6 — local/offline mode

- sqlite-backed local kanban board
- no external api needed — works fully offline
- import/export to json
- sync to remote provider (optional, manual trigger)
- good for personal task management or air-gapped environments

## v0.7 — multi-agent + mcp

**agent profiles**
- named agent configs instead of just codex/claude toggle
- `aboard.toml` or `~/.config/aboard/agents.toml`:
  ```toml
  [agents.codex]
  command = ["codex", "exec", "{prompt}"]

  [agents.claude]
  command = ["claude", "-p", "{prompt}"]

  [agents.custom]
  command = ["my-agent", "--context", "{context}"]
  ```
- cycle through agents with `a`, pick by number, or `/agent name`

**mcp server**
- expose aboard as an mcp server so agents can query board state
- tools: `get_board`, `get_card`, `move_card`, `create_card`, etc.
- agents call aboard directly instead of needing `<action>` blocks

**agent memory**
- persist agent conversation per card (local sqlite)
- resume context when reopening a card

## v0.8 — config file + theming

**config file**
- `aboard.toml` replaces env-only config
- provider settings, agent configs, keybindings, theme
- env vars still work as overrides

**custom keybindings**
- remap any key in config
- support modifier combos (ctrl, alt, super where terminal allows)

**theming**
- built-in themes: dark (default), light, catppuccin, gruvbox, nord
- custom theme via config: colors for header, columns, drawer, prompt, badges
- `ABOARD_THEME=catppuccin` or in config file
