package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/codywilliamson/aboard/internal/agent"
	"github.com/codywilliamson/aboard/internal/trello"
)

type boardsLoadedMsg struct {
	boards []trello.Board
	err    error
}

type boardDataLoadedMsg struct {
	boardID   string
	boardName string
	lists     []trello.List
	cards     []trello.Card
	err       error
}

type agentResponseMsg struct {
	agent  agent.AgentName
	prompt string
	output string
	err    error
}

type cardMutatedMsg struct {
	action string
	cardID string
	err    error
}

type listMutatedMsg struct {
	action string
	listID string
	err    error
}

func loadBoardsCmd(client trelloClient) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		boards, err := client.Boards(ctx)
		return boardsLoadedMsg{boards: boards, err: err}
	}
}

func loadBoardDataCmd(client trelloClient, boardID, boardName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		lists, err := client.ListsForBoard(ctx, boardID)
		if err != nil {
			return boardDataLoadedMsg{boardID: boardID, boardName: boardName, err: err}
		}
		cards, err := client.CardsForBoard(ctx, boardID)
		if err != nil {
			return boardDataLoadedMsg{boardID: boardID, boardName: boardName, err: err}
		}
		return boardDataLoadedMsg{boardID: boardID, boardName: boardName, lists: lists, cards: cards}
	}
}

func askAgentCmd(runner agentRunner, active agent.AgentName, cardContext, prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		output, err := runner.Ask(ctx, active, cardContext, prompt)
		return agentResponseMsg{agent: active, prompt: prompt, output: output, err: err}
	}
}

func moveCardCmd(client trelloClient, cardID, listID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := client.MoveCard(ctx, cardID, listID)
		return cardMutatedMsg{action: "move", cardID: cardID, err: err}
	}
}

func updateCardCmd(client trelloClient, cardID, name, desc string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := client.UpdateCard(ctx, cardID, name, desc)
		return cardMutatedMsg{action: "rename", cardID: cardID, err: err}
	}
}

func addCommentCmd(client trelloClient, cardID, text string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := client.AddComment(ctx, cardID, text)
		return cardMutatedMsg{action: "comment", cardID: cardID, err: err}
	}
}

func archiveCardCmd(client trelloClient, cardID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := client.ArchiveCard(ctx, cardID)
		return cardMutatedMsg{action: "archive", cardID: cardID, err: err}
	}
}

func createCardCmd(client trelloClient, listID, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		card, err := client.CreateCard(ctx, listID, name)
		id := ""
		if card != nil {
			id = card.ID
		}
		return cardMutatedMsg{action: "create", cardID: id, err: err}
	}
}

func createListCmd(client trelloClient, boardID, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		list, err := client.CreateList(ctx, boardID, name)
		id := ""
		if list != nil {
			id = list.ID
		}
		return listMutatedMsg{action: "create", listID: id, err: err}
	}
}

func archiveListCmd(client trelloClient, listID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := client.ArchiveList(ctx, listID)
		return listMutatedMsg{action: "archive", listID: listID, err: err}
	}
}
