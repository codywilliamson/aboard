package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/codywilliamson/aboard/internal/trello"
)

type promptMode int

const (
	promptAgent promptMode = iota
	promptMove
	promptRename
	promptComment
	promptNewCard
	promptNewList
	promptConfirmArchiveCard
	promptConfirmArchiveList
)

type PromptBar struct {
	mode       promptMode
	input      textinput.Model
	focused    bool
	width      int
	lists      []trello.List
	listCursor int
	confirmLabel string
}

func NewPromptBar() PromptBar {
	ti := textinput.New()
	ti.Placeholder = "ask the agent..."
	ti.CharLimit = 1200
	ti.Width = 60
	return PromptBar{
		mode:  promptAgent,
		input: ti,
	}
}

func (p *PromptBar) SetMode(mode promptMode) {
	p.mode = mode
	switch mode {
	case promptAgent:
		p.input.Placeholder = "ask the agent..."
	case promptRename:
		p.input.Placeholder = "new name..."
	case promptComment:
		p.input.Placeholder = "add comment..."
	case promptNewCard:
		p.input.Placeholder = "new card name..."
	case promptNewList:
		p.input.Placeholder = "new list name..."
	}
}

func (p *PromptBar) SetMoveLists(lists []trello.List, currentIdx int) {
	p.mode = promptMove
	p.lists = lists
	p.listCursor = currentIdx
}

func (p *PromptBar) SetConfirmArchiveCard(label string) {
	p.mode = promptConfirmArchiveCard
	p.confirmLabel = label
}

func (p *PromptBar) SetConfirmArchiveList(label string) {
	p.mode = promptConfirmArchiveList
	p.confirmLabel = label
}

func (p *PromptBar) Prefill(text string) {
	p.input.SetValue(text)
	p.input.CursorEnd()
}

func (p *PromptBar) Focus() {
	p.focused = true
	if p.mode != promptMove && p.mode != promptConfirmArchiveCard && p.mode != promptConfirmArchiveList {
		p.input.Focus()
	}
}

func (p *PromptBar) Blur() {
	p.focused = false
	p.input.Blur()
}

func (p *PromptBar) Value() string {
	return p.input.Value()
}

func (p *PromptBar) Reset() {
	p.mode = promptAgent
	p.input.SetValue("")
	p.input.Placeholder = "ask the agent..."
	p.lists = nil
	p.listCursor = 0
	p.confirmLabel = ""
}

func (p *PromptBar) Resize(w int) {
	p.width = w
	p.input.Width = max(20, w-16)
}

func (p *PromptBar) MoveListLeft() {
	if p.listCursor > 0 {
		p.listCursor--
	}
}

func (p *PromptBar) MoveListRight() {
	if p.listCursor < len(p.lists)-1 {
		p.listCursor++
	}
}

func (p *PromptBar) SelectedListID() string {
	if p.listCursor < 0 || p.listCursor >= len(p.lists) {
		return ""
	}
	return p.lists[p.listCursor].ID
}

func (p *PromptBar) View() string {
	w := max(40, p.width-2)
	style := promptBarStyle.Width(w)
	if p.focused {
		style = promptBarFocusedStyle.Width(w)
	}

	var content string
	switch p.mode {
	case promptMove:
		content = p.renderMoveView()
	case promptConfirmArchiveCard:
		content = p.renderConfirmView("archive card")
	case promptConfirmArchiveList:
		content = p.renderConfirmView("archive list")
	default:
		badge := promptModeBadgeStyle.Render(p.modeLabel())
		content = badge + " " + p.input.View()
	}

	return style.Render(content)
}

func (p *PromptBar) modeLabel() string {
	switch p.mode {
	case promptAgent:
		return "agent"
	case promptRename:
		return "rename"
	case promptComment:
		return "comment"
	case promptNewCard:
		return "card"
	case promptNewList:
		return "list"
	default:
		return "prompt"
	}
}

func (p *PromptBar) renderMoveView() string {
	badge := promptModeBadgeStyle.Render("move")
	var parts []string
	for i, list := range p.lists {
		name := ellipsis(list.Name, 16)
		if i == p.listCursor {
			parts = append(parts, promptMoveSelectedStyle.Render("▸"+name+"◂"))
		} else {
			parts = append(parts, promptMoveNormalStyle.Render(name))
		}
	}
	picker := strings.Join(parts, " | ")
	hint := subtleStyle.Render("  enter: confirm  esc: cancel")
	return badge + " " + picker + hint
}

func (p *PromptBar) renderConfirmView(action string) string {
	badge := promptModeBadgeStyle.Render("archive")
	label := fmt.Sprintf("%s %q?", action, p.confirmLabel)
	hint := subtleStyle.Render("  y: confirm  n/esc: cancel")
	return badge + " " + label + hint
}
