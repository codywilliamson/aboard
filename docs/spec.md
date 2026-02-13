# Trello Agent TUI Spec

## 1. Product Summary

A cross-platform terminal application that lets users:

- browse Trello boards as kanban columns,
- choose a card as working context,
- collaborate with local coding agents (`codex`, `claude`) against that context,
- keep interaction and status history inside the same TUI surface.

Primary platforms: Linux (x86_64), Windows (x86_64)

## 2. Goals

- Kanban-first view: columns = Trello lists, cards inside each column.
- Minimize context switching between browser, terminal, and AI tools.
- Polished, intentional UX with vim-style navigation.
- Fast keyboard workflows with drawer-based card detail + agent interaction.

## 3. Non-goals (v1)

- Full Trello CRUD (create/edit/archive cards, drag/drop lists).
- Multi-user collaboration inside the TUI.
- Persisted local database.

## 4. Layout

### Drawer closed (full kanban)

```
┌─────────────────────────────────────────────┐
│ board name          agent:codex   status    │
├────────┬────────┬────────┬──────────────────┤
│ List 1 │ List 2 │ List 3 │ List 4           │
│>card a │ card d │ card g │                  │
│ card b │ card e │ card h │                  │
│ card c │ card f │        │                  │
├────────┴────────┴────────┴──────────────────┤
│ h/l: columns  j/k: cards  enter: open  ?   │
└─────────────────────────────────────────────┘
```

### Drawer open (~40% kanban / ~60% drawer)

```
┌─────────────────────────────────────────────┐
│ board name          agent:codex   status    │
├────────┬────────┬───────────────────────────┤
│ List 1 │ List 2 │ Card Detail               │
│>card a │ card d │ name, list, url, desc     │
│ card b │ card e │ ─── Prompt ───            │
│ card c │ card f │ [input field]             │
│        │        │ ─── Timeline ───          │
│        │        │ [scrollable output]       │
├────────┴────────┴───────────────────────────┤
│ tab: focus  esc: close  enter: send         │
└─────────────────────────────────────────────┘
```

## 5. Navigation

| Context | Key | Action |
|---------|-----|--------|
| Kanban | h/l / ←/→ | Move between columns |
| Kanban | j/k / ↑/↓ | Move within column |
| Kanban | enter | Set context card + open drawer |
| Kanban | 1/2/a | Agent controls |
| Kanban | r/b | Refresh / board selector |
| Kanban | q | Quit |
| Drawer | enter/ctrl+s | Send prompt |
| Drawer | tab | Focus back to kanban |
| Drawer | esc | Close drawer |
| Drawer | ctrl+j/k | Scroll timeline |
| Global | ctrl+c | Quit |
| Global | ctrl+a/r/b | Toggle agent / refresh / boards |
| Global | ? | Toggle help overlay |

## 6. Architecture

```
internal/ui/
├── model.go      # root model, mode routing, global keys, header/status
├── messages.go   # tea.Msg types + tea.Cmd constructors
├── styles.go     # lipgloss styles, badge helper, color palette
├── boards.go     # board selector (render + update)
├── kanban.go     # KanbanModel: columns, per-list cursors, h/j/k/l nav
├── drawer.go     # DrawerModel: card detail + prompt input + timeline
├── help.go       # HelpModel: modal overlay with keybindings
└── util.go       # ellipsis, shortID, wrapForPane, clipToLineCount
```

### Key data models

**KanbanModel**: lists, cards grouped by list ID, list cursor, per-list card cursors, horizontal scroll offset, context card.

**DrawerModel**: current card, text input, scrollable viewport timeline, timeline entries.

### Data flow

1. Init → load boards or board data (lists + cards combined)
2. Board select → pick board → load board data
3. Kanban mode → navigate columns/cards → enter opens drawer
4. Drawer → type prompt → send → agent response in timeline

## 7. Functional Requirements

### FR-1 Auth and Boot

- Load config from env and optional `.env` file.
- Show setup screen if Trello auth vars are missing.

### FR-2 Board Selection

- Load open boards for member.
- Navigate with j/k, open with enter.

### FR-3 Kanban View

- Columns = Trello lists in board order.
- Cards displayed in each column.
- Auto-size columns: `terminal_width / 24` visible, min 1.
- Horizontal scrolling when more lists than visible columns.
- Active column and cursor card highlighted.

### FR-4 Drawer

- Card detail: name, list, URL, description.
- Prompt input for agent interaction.
- Scrollable timeline of system/user/agent entries.
- Opens on enter, closes on esc.

### FR-5 Agent Commands

- Support `codex` and `claude` via env-configured commands.
- `{prompt}` and `{context}` placeholders; stdin fallback.
- Timeline logs prompts immediately, responses on completion.
- On failure, restore prompt input for retry.

### FR-6 Feedback and Status

- Header shows board name, active agent badge, status.
- Footer shows context-sensitive keybindings.

## 8. Acceptance Criteria

- AC-1: Board selector appears when no board ID configured.
- AC-2: Selecting board shows kanban columns with cards by list.
- AC-3: h/l navigates columns, j/k navigates cards within column.
- AC-4: Enter opens drawer with card detail + prompt.
- AC-5: Sending prompt creates immediate user timeline entry.
- AC-6: Agent response appears in timeline; errors restore prompt.
- AC-7: Esc closes drawer, ? shows help overlay.
- AC-8: Terminal resize works without crash.
- AC-9: `go build ./cmd/trello-tui` and `go vet ./...` clean.
