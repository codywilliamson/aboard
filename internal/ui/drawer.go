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
	hadCard := d.card != nil
	d.card = card
	hasCard := d.card != nil
	if hadCard != hasCard && d.width > 0 {
		d.Resize(d.width, d.height)
	}
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
	// reserve space for card detail when a card is open, otherwise just timeline header + borders
	overhead := 4
	if d.card != nil {
		overhead = 12
	}
	d.timeline.Height = max(3, h-overhead)
}

func (d *DrawerModel) View() string {
	innerWidth := max(20, d.width-4)

	var sections []string
	if d.card != nil {
		sections = append(sections, d.renderCardDetail(innerWidth), "")
	}

	timelineTitle := lipgloss.NewStyle().Bold(true).Render("─── Timeline ───")
	var timelineContent string
	if len(d.entries) == 0 {
		timelineContent = subtleStyle.Render("No output yet. Send a prompt with / or tab.")
	} else {
		timelineContent = d.timeline.View()
	}
	sections = append(sections, timelineTitle, timelineContent)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
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
	desc = stripMarkdown(desc)
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
