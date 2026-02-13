package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/codywilliamson/aboard/internal/trello"
)

type timelineEntry struct {
	stamp string
	who   string
	text  string
}

type DrawerModel struct {
	card     *trello.Card
	timeline viewport.Model
	entries  []timelineEntry
	width    int
	height   int
}

func NewDrawerModel() DrawerModel {
	return DrawerModel{
		timeline: viewport.New(60, 10),
	}
}

func (d *DrawerModel) SetCard(card *trello.Card) {
	d.card = card
}

func (d *DrawerModel) AppendTimeline(who, text string) {
	stamp := time.Now().Format("15:04:05")
	d.entries = append(d.entries, timelineEntry{stamp: stamp, who: who, text: strings.TrimSpace(text)})
	d.rebuildTimeline()
}

func (d *DrawerModel) rebuildTimeline() {
	var parts []string
	for _, e := range d.entries {
		parts = append(parts, fmt.Sprintf("[%s] %s\n%s", e.stamp, e.who, e.text))
	}
	content := strings.Join(parts, "\n\n")
	d.timeline.SetContent(content)
	d.timeline.GotoBottom()
}

func (d *DrawerModel) Resize(w, h int) {
	d.width = w
	d.height = h
	d.timeline.Width = max(20, w-4)
	d.timeline.Height = max(3, h-12)
}

func (d *DrawerModel) View() string {
	if d.card == nil {
		return drawerStyle.Width(d.width).Height(d.height).Render(
			subtleStyle.Render("No card selected."),
		)
	}

	innerWidth := max(20, d.width-4)
	detail := d.renderCardDetail(innerWidth)

	timelineTitle := lipgloss.NewStyle().Bold(true).Render("─── Timeline ───")
	var timelineContent string
	if len(d.entries) == 0 {
		timelineContent = subtleStyle.Render("No output yet. Send a prompt with / or tab.")
	} else {
		timelineContent = d.timeline.View()
	}
	timeline := lipgloss.JoinVertical(lipgloss.Left, timelineTitle, timelineContent)

	content := lipgloss.JoinVertical(lipgloss.Left, detail, "", timeline)
	return drawerStyle.Width(d.width).Height(d.height).Render(content)
}

func (d *DrawerModel) renderCardDetail(width int) string {
	card := d.card
	if card == nil {
		return ""
	}

	url := card.ShortURL
	if url == "" {
		url = card.URL
	}
	if url == "" {
		url = "(no url)"
	}

	desc := strings.TrimSpace(card.Desc)
	if desc == "" {
		desc = "(no description)"
	}
	desc = wrapForPane(desc, max(20, width-4))

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render(ellipsis(card.Name, width-2))

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Card Detail"),
		title,
		"list: "+listBadge(card.ListName),
		subtleStyle.Render("url: "+url),
		"",
		subtleStyle.Render(desc),
	)
}
