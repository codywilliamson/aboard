package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/codywilliamson/aboard/internal/trello"
)

type KanbanModel struct {
	lists        []trello.List
	cards        map[string][]trello.Card
	listCursor   int
	cardCursors  map[string]int
	scrollOffset int
	contextCard  *trello.Card
	width        int
	height       int
}

func (k *KanbanModel) SetData(lists []trello.List, cards []trello.Card) {
	k.lists = lists
	k.cards = make(map[string][]trello.Card, len(lists))
	for _, card := range cards {
		k.cards[card.IDList] = append(k.cards[card.IDList], card)
	}
	k.listCursor = 0
	k.scrollOffset = 0
	k.cardCursors = make(map[string]int, len(lists))
	k.contextCard = nil
}

func (k *KanbanModel) activeListID() string {
	if k.listCursor < 0 || k.listCursor >= len(k.lists) {
		return ""
	}
	return k.lists[k.listCursor].ID
}

func (k *KanbanModel) ActiveList() *trello.List {
	if k.listCursor < 0 || k.listCursor >= len(k.lists) {
		return nil
	}
	l := k.lists[k.listCursor]
	return &l
}

func (k *KanbanModel) SelectedCard() *trello.Card {
	id := k.activeListID()
	if id == "" {
		return nil
	}
	cards := k.cards[id]
	cursor := k.cardCursors[id]
	if cursor < 0 || cursor >= len(cards) {
		return nil
	}
	c := cards[cursor]
	return &c
}

func (k *KanbanModel) MovePrevList() {
	if k.listCursor > 0 {
		k.listCursor--
		k.ensureHorizontalScroll()
	}
}

func (k *KanbanModel) MoveNextList() {
	if k.listCursor < len(k.lists)-1 {
		k.listCursor++
		k.ensureHorizontalScroll()
	}
}

func (k *KanbanModel) MovePrevCard() {
	id := k.activeListID()
	if id == "" {
		return
	}
	if k.cardCursors[id] > 0 {
		k.cardCursors[id]--
	}
}

func (k *KanbanModel) MoveNextCard() {
	id := k.activeListID()
	if id == "" {
		return
	}
	if k.cardCursors[id] < len(k.cards[id])-1 {
		k.cardCursors[id]++
	}
}

func (k *KanbanModel) visibleColumns() int {
	if k.width <= 0 {
		return 1
	}
	return max(1, k.width/24)
}

func (k *KanbanModel) ensureHorizontalScroll() {
	vis := k.visibleColumns()
	if k.listCursor < k.scrollOffset {
		k.scrollOffset = k.listCursor
	}
	if k.listCursor >= k.scrollOffset+vis {
		k.scrollOffset = k.listCursor - vis + 1
	}
}

func (k *KanbanModel) View() string {
	if len(k.lists) == 0 {
		return subtleStyle.Render("  no lists to display.")
	}

	vis := k.visibleColumns()
	colWidth := max(16, (k.width-2)/vis-2)
	colHeight := max(4, k.height-2)

	startIdx := k.scrollOffset
	endIdx := min(startIdx+vis, len(k.lists))

	var columns []string
	for i := startIdx; i < endIdx; i++ {
		list := k.lists[i]
		col := k.renderColumn(list, colWidth, colHeight, i == k.listCursor)
		columns = append(columns, col)
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	if len(k.lists) > vis {
		indicator := subtleStyle.Render(fmt.Sprintf(" %d/%d lists (h/l to scroll)", k.listCursor+1, len(k.lists)))
		return lipgloss.JoinVertical(lipgloss.Left, board, indicator)
	}
	return board
}

func (k *KanbanModel) renderColumn(list trello.List, width, height int, active bool) string {
	cards := k.cards[list.ID]
	cursor := k.cardCursors[list.ID]

	header := lipgloss.NewStyle().Bold(true).Render(ellipsis(list.Name, width-2))
	noun := "cards"
	if len(cards) == 1 {
		noun = "card"
	}
	countStr := subtleStyle.Render(fmt.Sprintf("%d %s", len(cards), noun))

	// visible card range — ensure cursor stays in view
	maxCards := max(1, height-4)
	startIdx := 0
	if active && cursor >= maxCards {
		startIdx = cursor - maxCards + 1
	}
	endIdx := min(startIdx+maxCards, len(cards))

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		card := cards[i]
		prefix := "  "
		if active && i == cursor {
			prefix = "> "
		}

		isContext := k.contextCard != nil && k.contextCard.ID == card.ID
		if isContext {
			prefix = string(prefix[0:1]) + "◆"
		}

		name := ellipsis(card.Name, width-4)
		line := prefix + name

		if active && i == cursor {
			line = selectedRowStyle.Render(line)
		} else if isContext {
			line = contextMarkerStyle.Render(line)
		}
		lines = append(lines, line)
	}

	if len(cards) == 0 {
		lines = append(lines, subtleStyle.Render("  (empty)"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		countStr,
		strings.Join(lines, "\n"),
	)

	style := columnStyle
	if active {
		style = activeColumnStyle
	}
	return style.Width(width).Height(height).Render(content)
}
