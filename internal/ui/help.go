package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HelpModel struct {
	visible bool
}

func (h *HelpModel) Render(bg string, w, ht int) string {
	help := strings.Join([]string{
		"Keymap",
		"",
		"Kanban",
		"  h/l  ←/→   move between columns",
		"  j/k  ↑/↓   move within column",
		"  enter       open card in drawer",
		"  m           move card (list picker)",
		"  e           rename card",
		"  c           comment on card",
		"  n           new card in current list",
		"  N           new list on board",
		"  x           archive card (confirm)",
		"  X           archive list (confirm)",
		"  / or tab    focus prompt bar",
		"  1/2/a       agent controls",
		"  r           refresh board data",
		"  b           board selector",
		"  q           quit",
		"",
		"Prompt",
		"  enter       submit (context-dependent)",
		"  esc         cancel, return to kanban",
		"  h/l         navigate list picker (move)",
		"",
		"Drawer",
		"  j/k         scroll timeline",
		"  tab         focus kanban",
		"  /           focus prompt",
		"  m/e/c/x     card operations",
		"",
		"Global",
		"  ctrl+c      quit",
		"  ctrl+a      toggle agent",
		"  ctrl+r      refresh",
		"  ctrl+b      board selector",
		"  ?           toggle help",
		"",
		"Press ? or Esc to close.",
	}, "\n")

	panel := helpPanelStyle.Width(max(50, w-16)).Render(help)
	return lipgloss.Place(w, ht, lipgloss.Center, lipgloss.Center, panel,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("238")),
	)
}
