# architecture

## current state (v0.1)

### what works
- kanban view: columns = lists, per-list cursors, horizontal scroll
- card drawer: detail panel + scrollable agent timeline
- global prompt bar: agent, move, rename, comment, new card, new list, archive
- agent integration: codex + claude via configurable CLI commands
- agent actions: structured `<action>` blocks in agent responses trigger board mutations
- trello api: full read + write (move, rename, comment, archive, create cards/lists)
- ci + goreleaser cross-platform releases

### file layout

```
cmd/aboard/          entry point
internal/
  agent/runner.go    agent CLI execution + prompt building
  config/config.go   env + .env loading
  trello/client.go   trello api client (read + write)
  ui/
    model.go         root model, routing, focus management
    messages.go      tea messages + mutation commands
    kanban.go        kanban columns + navigation
    drawer.go        card detail + timeline
    prompt.go        multi-mode prompt bar
    actions.go       agent action parser + executor
    boards.go        board selector
    help.go          help overlay
    styles.go        lipgloss styles
    util.go          text helpers
```

### data flow

1. **init** → load boards or board data (lists + cards)
2. **board select** → pick board → load board data
3. **kanban mode** → navigate columns/cards → enter opens drawer
4. **drawer** → card detail + scrollable timeline
5. **prompt bar** → multi-mode input (agent, move, rename, comment, create, archive)
6. **agent** → send prompt with board context → response in timeline → parse `<action>` blocks → execute mutations → auto-refresh

### focus system

three focus areas: `kanban`, `drawer`, `prompt`

- kanban: column/card navigation, shortcut keys trigger prompt modes
- drawer: timeline scrolling, delegates card ops to prompt
- prompt: text input or list picker depending on mode, esc cascades back

### mutation cycle

```
user action or agent <action> block
  → mutation command (tea.Cmd)
    → trello api call
      → cardMutatedMsg / listMutatedMsg
        → auto-refresh board data
```

### agent action protocol

agents include structured blocks in their response:

```
<action>{"type":"move_card","card_id":"...","list_id":"..."}</action>
```

the parser strips action blocks from display text, executes them as mutation commands via `tea.Batch`, and surfaces api errors in the timeline.

supported types: `move_card`, `update_card`, `add_comment`, `archive_card`, `create_card`, `create_list`, `archive_list`
