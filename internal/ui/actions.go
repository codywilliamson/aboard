package ui

import (
	"encoding/json"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type agentAction struct {
	Type   string `json:"type"`
	CardID string `json:"card_id,omitempty"`
	ListID string `json:"list_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Desc   string `json:"desc,omitempty"`
	Text   string `json:"text,omitempty"`
}

var actionRe = regexp.MustCompile(`<action>(.*?)</action>`)

// parseActions extracts <action>{...}</action> blocks from agent output.
// returns cleaned display text and parsed actions.
func parseActions(raw string) (string, []agentAction) {
	matches := actionRe.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return raw, nil
	}

	display := actionRe.ReplaceAllString(raw, "")
	display = strings.TrimSpace(display)

	var actions []agentAction
	for _, match := range matches {
		var a agentAction
		if err := json.Unmarshal([]byte(match[1]), &a); err != nil {
			continue
		}
		if a.Type == "" {
			continue
		}
		actions = append(actions, a)
	}
	return display, actions
}

// executeActions converts parsed actions into mutation commands.
func executeActions(client trelloClient, boardID string, actions []agentAction) tea.Cmd {
	var cmds []tea.Cmd
	for _, a := range actions {
		switch a.Type {
		case "move_card":
			if a.CardID != "" && a.ListID != "" {
				cmds = append(cmds, moveCardCmd(client, a.CardID, a.ListID))
			}
		case "update_card":
			if a.CardID != "" && (a.Name != "" || a.Desc != "") {
				cmds = append(cmds, updateCardCmd(client, a.CardID, a.Name, a.Desc))
			}
		case "add_comment":
			if a.CardID != "" && a.Text != "" {
				cmds = append(cmds, addCommentCmd(client, a.CardID, a.Text))
			}
		case "archive_card":
			if a.CardID != "" {
				cmds = append(cmds, archiveCardCmd(client, a.CardID))
			}
		case "create_card":
			if a.ListID != "" && a.Name != "" {
				cmds = append(cmds, createCardCmd(client, a.ListID, a.Name))
			}
		case "create_list":
			if a.Name != "" {
				cmds = append(cmds, createListCmd(client, boardID, a.Name))
			}
		case "archive_list":
			if a.ListID != "" {
				cmds = append(cmds, archiveListCmd(client, a.ListID))
			}
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}
