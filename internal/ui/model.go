package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/codywilliamson/aboard/internal/agent"
	"github.com/codywilliamson/aboard/internal/config"
	"github.com/codywilliamson/aboard/internal/trello"
)

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

type agentRunner interface {
	Ask(context.Context, agent.AgentName, string, string) (string, error)
}

type viewMode int

const (
	modeBoardSelect viewMode = iota
	modeKanban
)

type focusArea int

const (
	focusKanban focusArea = iota
	focusDrawer
	focusPrompt
)

type Model struct {
	cfg    config.Config
	trello trelloClient
	runner agentRunner

	mode       viewMode
	focus      focusArea
	active     agent.AgentName
	status     string
	errText    string
	loading    bool
	runningAsk bool

	width  int
	height int

	boards      []trello.Board
	boardCursor int
	boardID     string
	boardName   string

	kanban     KanbanModel
	drawer     DrawerModel
	prompt     PromptBar
	help       HelpModel
	drawerOpen bool

	// operation targets for prompt actions
	opCardID string
	opListID string

	pendingPrompt string
}

func NewModel(cfg config.Config, tc trelloClient, runner agentRunner) Model {
	return Model{
		cfg:     cfg,
		trello:  tc,
		runner:  runner,
		mode:    modeKanban,
		focus:   focusKanban,
		active:  agent.AgentCodex,
		status:  "loading...",
		boardID: cfg.TrelloBoardID,
		drawer:  NewDrawerModel(),
		prompt:  NewPromptBar(),
		kanban: KanbanModel{
			cardCursors: make(map[string]int),
		},
	}
}

func (m Model) Init() tea.Cmd {
	if !m.trello.CanAuth() {
		return nil
	}
	if m.boardID != "" {
		return loadBoardDataCmd(m.trello, m.boardID, "")
	}
	return loadBoardsCmd(m.trello)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()
		return m, nil

	case boardsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errText = msg.err.Error()
			m.status = "failed to load boards"
			return m, nil
		}
		m.boards = msg.boards
		m.mode = modeBoardSelect
		m.errText = ""
		if len(m.boards) == 0 {
			m.status = "no open boards found"
		} else {
			m.status = "select a board"
		}
		return m, nil

	case boardDataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errText = msg.err.Error()
			m.status = "failed to load board"
			return m, nil
		}
		m.boardID = msg.boardID
		if msg.boardName != "" {
			m.boardName = msg.boardName
		} else {
			m.boardName = m.boardNameByID(msg.boardID)
		}
		m.errText = ""
		m.mode = modeKanban
		m.kanban.SetData(msg.lists, msg.cards)
		m.recalcLayout()
		if len(msg.lists) == 0 {
			m.status = "board loaded (no lists)"
		} else {
			m.status = "h/l: columns  j/k: cards  enter: open  ?: help"
		}
		return m, nil

	case agentResponseMsg:
		m.runningAsk = false
		if msg.err != nil {
			m.errText = msg.err.Error()
			m.status = "agent failed"
			if strings.TrimSpace(m.pendingPrompt) != "" && strings.TrimSpace(m.prompt.Value()) == "" {
				m.prompt.input.SetValue(m.pendingPrompt)
			}
			m.drawer.AppendTimeline(string(msg.agent)+" error", msg.err.Error())
			m.pendingPrompt = ""
			return m, nil
		}
		m.errText = ""
		m.status = fmt.Sprintf("%s responded", msg.agent)
		displayText, actions := parseActions(msg.output)
		m.drawer.AppendTimeline(string(msg.agent), displayText)
		m.pendingPrompt = ""
		if len(actions) > 0 {
			m.status = fmt.Sprintf("executing %d action(s)...", len(actions))
			return m, executeActions(m.trello, m.boardID, actions)
		}
		return m, nil

	case cardMutatedMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			m.status = fmt.Sprintf("%s failed", msg.action)
			m.drawer.AppendTimeline("system", fmt.Sprintf("%s failed: %s", msg.action, msg.err.Error()))
			return m, nil
		}
		m.errText = ""
		m.status = fmt.Sprintf("card %s ok", msg.action)
		m.drawer.AppendTimeline("system", fmt.Sprintf("card %s ok", msg.action))
		return m, loadBoardDataCmd(m.trello, m.boardID, m.boardName)

	case listMutatedMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			m.status = fmt.Sprintf("list %s failed", msg.action)
			m.drawer.AppendTimeline("system", fmt.Sprintf("list %s failed: %s", msg.action, msg.err.Error()))
			return m, nil
		}
		m.errText = ""
		m.status = fmt.Sprintf("list %s ok", msg.action)
		m.drawer.AppendTimeline("system", fmt.Sprintf("list %s ok", msg.action))
		return m, loadBoardDataCmd(m.trello, m.boardID, m.boardName)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.help.visible {
		if key == "?" || key == "esc" {
			m.help.visible = false
		}
		return m, nil
	}

	// global keys
	switch key {
	case "?":
		m.help.visible = true
		return m, nil
	case "ctrl+a":
		m.toggleAgent()
		return m, nil
	case "ctrl+r":
		cmd := m.refreshData()
		return m, cmd
	case "ctrl+b":
		cmd := m.openBoardSelector()
		return m, cmd
	}

	if m.mode == modeBoardSelect {
		return m.updateBoardSelect(msg)
	}

	// kanban mode — route by focus
	switch m.focus {
	case focusPrompt:
		return m.updatePromptKeys(msg)
	case focusDrawer:
		return m.updateDrawerKeys(msg)
	default:
		return m.updateKanbanKeys(msg)
	}
}

func (m Model) updateKanbanKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "h", "left":
		m.kanban.MovePrevList()
	case "l", "right":
		m.kanban.MoveNextList()
	case "j", "down":
		m.kanban.MoveNextCard()
	case "k", "up":
		m.kanban.MovePrevCard()
	case "enter":
		if card := m.kanban.SelectedCard(); card != nil {
			m.kanban.contextCard = card
			m.drawer.SetCard(card)
			m.drawerOpen = true
			m.focus = focusDrawer
			m.recalcLayout()
			m.status = "drawer open — tab: switch focus  esc: close"
		}
	case "m":
		return m.startMoveCard()
	case "e":
		return m.startRenameCard()
	case "c":
		return m.startCommentCard()
	case "n":
		return m.startNewCard()
	case "N":
		return m.startNewList()
	case "x":
		return m.startArchiveCard()
	case "X":
		return m.startArchiveList()
	case "/", "tab":
		m.focusPromptBar(promptAgent)
	case "q":
		return m, tea.Quit
	case "1":
		m.setAgent(agent.AgentCodex)
	case "2":
		m.setAgent(agent.AgentClaude)
	case "a":
		m.toggleAgent()
	case "r":
		cmd := m.refreshData()
		return m, cmd
	case "b":
		cmd := m.openBoardSelector()
		return m, cmd
	}
	return m, nil
}

func (m Model) updateDrawerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.drawerOpen = false
		m.focus = focusKanban
		m.recalcLayout()
		m.status = "h/l: columns  j/k: cards  enter: open  ?: help"
		return m, nil
	case "tab":
		m.focus = focusKanban
		return m, nil
	case "j", "down":
		m.drawer.timeline.LineDown(3)
		return m, nil
	case "k", "up":
		m.drawer.timeline.LineUp(3)
		return m, nil
	case "/":
		m.focusPromptBar(promptAgent)
		return m, nil
	case "m":
		return m.startMoveCard()
	case "e":
		return m.startRenameCard()
	case "c":
		return m.startCommentCard()
	case "x":
		return m.startArchiveCard()
	}
	return m, nil
}

func (m Model) updatePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// tab exits prompt in all modes
	if key == "tab" {
		m.cancelPrompt()
		return m, nil
	}

	switch m.prompt.mode {
	case promptMove:
		switch key {
		case "h", "left":
			m.prompt.MoveListLeft()
		case "l", "right":
			m.prompt.MoveListRight()
		case "enter":
			return m.submitMove()
		case "esc":
			m.cancelPrompt()
		}
		return m, nil

	case promptConfirmArchiveCard:
		switch key {
		case "y":
			return m.submitArchiveCard()
		case "n", "esc":
			m.cancelPrompt()
		}
		return m, nil

	case promptConfirmArchiveList:
		switch key {
		case "y":
			return m.submitArchiveList()
		case "n", "esc":
			m.cancelPrompt()
		}
		return m, nil

	default:
		// text input modes: agent, rename, comment, new card, new list
		switch key {
		case "esc":
			m.cancelPrompt()
			return m, nil
		case "enter":
			return m.submitPrompt()
		}
		var cmd tea.Cmd
		m.prompt.input, cmd = m.prompt.input.Update(msg)
		return m, cmd
	}
}

// --- prompt mode starters ---

func (m *Model) focusPromptBar(mode promptMode) {
	m.prompt.SetMode(mode)
	m.focus = focusPrompt
	m.prompt.Focus()
}

func (m Model) startMoveCard() (tea.Model, tea.Cmd) {
	card := m.kanban.SelectedCard()
	if card == nil {
		m.status = "no card selected"
		return m, nil
	}
	m.opCardID = card.ID
	m.prompt.SetMoveLists(m.kanban.lists, m.kanban.listCursor)
	m.focus = focusPrompt
	m.prompt.Focus()
	m.status = "h/l: pick list  enter: confirm  esc: cancel"
	return m, nil
}

func (m Model) startRenameCard() (tea.Model, tea.Cmd) {
	card := m.kanban.SelectedCard()
	if card == nil {
		m.status = "no card selected"
		return m, nil
	}
	m.opCardID = card.ID
	m.focusPromptBar(promptRename)
	m.prompt.Prefill(card.Name)
	m.status = "enter: rename  esc: cancel"
	return m, nil
}

func (m Model) startCommentCard() (tea.Model, tea.Cmd) {
	card := m.kanban.SelectedCard()
	if card == nil {
		m.status = "no card selected"
		return m, nil
	}
	m.opCardID = card.ID
	m.focusPromptBar(promptComment)
	m.status = "enter: add comment  esc: cancel"
	return m, nil
}

func (m Model) startNewCard() (tea.Model, tea.Cmd) {
	list := m.kanban.ActiveList()
	if list == nil {
		m.status = "no list selected"
		return m, nil
	}
	m.opListID = list.ID
	m.focusPromptBar(promptNewCard)
	m.status = "enter: create card  esc: cancel"
	return m, nil
}

func (m Model) startNewList() (tea.Model, tea.Cmd) {
	m.focusPromptBar(promptNewList)
	m.status = "enter: create list  esc: cancel"
	return m, nil
}

func (m Model) startArchiveCard() (tea.Model, tea.Cmd) {
	card := m.kanban.SelectedCard()
	if card == nil {
		m.status = "no card selected"
		return m, nil
	}
	m.opCardID = card.ID
	m.prompt.SetConfirmArchiveCard(ellipsis(card.Name, 30))
	m.focus = focusPrompt
	m.prompt.Focus()
	m.status = "y: archive  n/esc: cancel"
	return m, nil
}

func (m Model) startArchiveList() (tea.Model, tea.Cmd) {
	list := m.kanban.ActiveList()
	if list == nil {
		m.status = "no list selected"
		return m, nil
	}
	m.opListID = list.ID
	m.prompt.SetConfirmArchiveList(ellipsis(list.Name, 30))
	m.focus = focusPrompt
	m.prompt.Focus()
	m.status = "y: archive list  n/esc: cancel"
	return m, nil
}

// --- prompt submissions ---

func (m Model) submitPrompt() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.prompt.Value())
	if value == "" {
		m.status = "input is empty"
		return m, nil
	}

	switch m.prompt.mode {
	case promptAgent:
		return m.sendAgentPrompt(value)
	case promptRename:
		m.cancelPrompt()
		m.status = "renaming card..."
		return m, updateCardCmd(m.trello, m.opCardID, value, "")
	case promptComment:
		m.cancelPrompt()
		m.status = "adding comment..."
		return m, addCommentCmd(m.trello, m.opCardID, value)
	case promptNewCard:
		m.cancelPrompt()
		m.status = "creating card..."
		return m, createCardCmd(m.trello, m.opListID, value)
	case promptNewList:
		m.cancelPrompt()
		m.status = "creating list..."
		return m, createListCmd(m.trello, m.boardID, value)
	}
	m.cancelPrompt()
	return m, nil
}

func (m Model) submitMove() (tea.Model, tea.Cmd) {
	listID := m.prompt.SelectedListID()
	cardID := m.opCardID
	m.cancelPrompt()
	if listID == "" || cardID == "" {
		m.status = "move cancelled"
		return m, nil
	}
	m.status = "moving card..."
	return m, moveCardCmd(m.trello, cardID, listID)
}

func (m Model) submitArchiveCard() (tea.Model, tea.Cmd) {
	cardID := m.opCardID
	m.cancelPrompt()
	if cardID == "" {
		return m, nil
	}
	m.status = "archiving card..."
	return m, archiveCardCmd(m.trello, cardID)
}

func (m Model) submitArchiveList() (tea.Model, tea.Cmd) {
	listID := m.opListID
	m.cancelPrompt()
	if listID == "" {
		return m, nil
	}
	m.status = "archiving list..."
	return m, archiveListCmd(m.trello, listID)
}

func (m *Model) cancelPrompt() {
	m.prompt.Blur()
	m.prompt.Reset()
	m.opCardID = ""
	m.opListID = ""
	if m.drawerOpen {
		m.focus = focusDrawer
	} else {
		m.focus = focusKanban
	}
	m.status = "h/l: columns  j/k: cards  enter: open  ?: help"
}

func (m Model) sendAgentPrompt(prompt string) (tea.Model, tea.Cmd) {
	if m.runningAsk {
		m.status = "agent request in progress..."
		return m, nil
	}

	boardContext := m.currentBoardContext()
	m.runningAsk = true
	m.status = fmt.Sprintf("running %s...", m.active)
	m.pendingPrompt = prompt
	m.drawer.AppendTimeline("you", prompt)
	m.prompt.input.SetValue("")

	// open drawer if not already open
	if !m.drawerOpen {
		m.drawerOpen = true
		m.recalcLayout()
	}
	m.focus = focusDrawer

	return m, askAgentCmd(m.runner, m.active, boardContext, prompt)
}

func (m *Model) currentBoardContext() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Board: %s (id: %s)", m.boardName, m.boardID))

	// card context if available
	card := m.kanban.contextCard
	if card == nil {
		card = m.kanban.SelectedCard()
	}
	if card != nil {
		desc := strings.TrimSpace(card.Desc)
		if desc == "" {
			desc = "(no description)"
		}
		parts = append(parts,
			"",
			fmt.Sprintf("Selected Card: %s (id: %s)", card.Name, card.ID),
			fmt.Sprintf("List: %s", card.ListName),
			fmt.Sprintf("URL: %s", card.ShortURL),
			"Description:",
			desc,
		)
	}

	// all lists with IDs and card counts
	parts = append(parts, "", "Lists:")
	for _, list := range m.kanban.lists {
		cards := m.kanban.cards[list.ID]
		parts = append(parts, fmt.Sprintf("  - %s (id: %s, %d cards)", list.Name, list.ID, len(cards)))
	}

	return strings.Join(parts, "\n")
}

// --- view ---

func (m Model) View() string {
	if !m.trello.CanAuth() {
		return baseStyle.Render(strings.Join([]string{
			"trello auth not configured.",
			"",
			"set these env vars:",
			"  TRELLO_API_KEY",
			"  TRELLO_API_TOKEN",
			"  TRELLO_BOARD_ID (optional)",
		}, "\n"))
	}

	if m.mode == modeBoardSelect {
		content := m.renderBoardSelect()
		if m.help.visible {
			return m.help.Render(content, m.width, m.height)
		}
		return content
	}

	header := m.header()
	promptBar := m.prompt.View()

	var main string
	if m.drawerOpen {
		main = lipgloss.JoinHorizontal(lipgloss.Top,
			m.kanban.View(),
			m.drawer.View(),
		)
	} else {
		main = m.kanban.View()
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, main, promptBar)
	content = clipToLineCount(content, max(8, m.height))

	if m.help.visible {
		return m.help.Render(content, m.width, m.height)
	}
	return content
}

func (m *Model) header() string {
	left := m.headerLeft()
	right := m.headerRight()
	w := max(40, m.width-2)
	leftWidth := w / 2
	rightWidth := w - leftWidth

	return headerBoxStyle.Width(w).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(left),
			lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Right).Render(right),
		),
	)
}

func (m *Model) headerLeft() string {
	board := "board: none"
	if m.boardName != "" {
		board = "board: " + m.boardName
	} else if m.boardID != "" {
		board = "board: " + shortID(m.boardID)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		heroStyle.Render("aboard"),
		subtleStyle.Render(board),
	)
}

func (m *Model) headerRight() string {
	agentLabel := "agent:" + string(m.active)
	status := m.status
	if m.errText != "" {
		status = m.errText
	}
	return lipgloss.JoinVertical(lipgloss.Right,
		statusStyle.Render(agentLabel),
		subtleStyle.Render(status),
	)
}

// --- layout ---

func (m *Model) recalcLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	headerHeight := 4
	promptHeight := 3
	mainHeight := max(6, m.height-headerHeight-promptHeight-2)

	m.prompt.Resize(m.width)

	if m.drawerOpen {
		kanbanWidth := max(30, int(float64(m.width)*0.4))
		drawerWidth := max(30, m.width-kanbanWidth)
		m.kanban.width = kanbanWidth
		m.kanban.height = mainHeight
		m.drawer.Resize(drawerWidth, mainHeight)
	} else {
		m.kanban.width = m.width
		m.kanban.height = mainHeight
	}
}

// --- helpers ---

func (m *Model) toggleAgent() {
	if m.active == agent.AgentCodex {
		m.setAgent(agent.AgentClaude)
	} else {
		m.setAgent(agent.AgentCodex)
	}
}

func (m *Model) setAgent(next agent.AgentName) {
	if m.active == next {
		return
	}
	m.active = next
	m.status = fmt.Sprintf("agent: %s", m.active)
	m.drawer.AppendTimeline("system", m.status)
}

func (m *Model) refreshData() tea.Cmd {
	m.loading = true
	m.errText = ""

	if m.mode == modeBoardSelect {
		m.status = "refreshing boards..."
		return loadBoardsCmd(m.trello)
	}
	if m.boardID != "" {
		m.status = "refreshing board..."
		return loadBoardDataCmd(m.trello, m.boardID, m.boardName)
	}
	return m.openBoardSelector()
}

func (m *Model) openBoardSelector() tea.Cmd {
	if len(m.boards) > 0 {
		m.mode = modeBoardSelect
		m.status = "select a board"
		return nil
	}
	m.loading = true
	m.errText = ""
	m.mode = modeBoardSelect
	m.status = "loading boards..."
	return loadBoardsCmd(m.trello)
}

func (m *Model) boardNameByID(boardID string) string {
	for _, board := range m.boards {
		if board.ID == boardID {
			return board.Name
		}
	}
	return ""
}
