package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateBoardSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.boardCursor > 0 {
			m.boardCursor--
		}
	case "down", "j":
		if m.boardCursor < len(m.boards)-1 {
			m.boardCursor++
		}
	case "enter":
		if len(m.boards) > 0 {
			selected := m.boards[m.boardCursor]
			m.loading = true
			m.status = fmt.Sprintf("loading %q...", selected.Name)
			return m, loadBoardDataCmd(m.trello, selected.ID, selected.Name)
		}
	case "q":
		return m, tea.Quit
	case "r":
		cmd := m.refreshData()
		return m, cmd
	}
	return m, nil
}

func (m *Model) renderBoardSelect() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(heroStyle.Render("Trello Agent Console"))
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("Pick a board to begin."))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(infoStyle.Render("Loading boards..."))
		b.WriteString("\n")
	}
	if m.errText != "" {
		b.WriteString(errorStyle.Render(m.errText))
		b.WriteString("\n\n")
	}

	if len(m.boards) == 0 && !m.loading {
		b.WriteString(subtleStyle.Render("No boards to display. Press r to retry."))
		return baseStyle.Render(b.String())
	}

	for i, board := range m.boards {
		cursor := "  "
		if i == m.boardCursor {
			cursor = "▸ "
		}
		line := fmt.Sprintf("%s%s  %s", cursor, board.Name, subtleStyle.Render("["+shortID(board.ID)+"]"))
		if i == m.boardCursor {
			line = focusedBoardStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("enter: open   j/k: navigate   r: refresh   ?: help   q: quit"))

	return baseStyle.Render(b.String())
}
