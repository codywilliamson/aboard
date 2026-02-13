package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/codywilliamson/aboard/internal/agent"
	"github.com/codywilliamson/aboard/internal/config"
	"github.com/codywilliamson/aboard/internal/trello"
	"github.com/codywilliamson/aboard/internal/ui"
)

func main() {
	cfg := config.LoadFromEnv()

	trelloClient := trello.NewClient(cfg.TrelloAPIKey, cfg.TrelloAPIToken)
	runner := agent.NewRunner(cfg.CodexCommand, cfg.ClaudeCommand)

	m := ui.NewModel(cfg, trelloClient, runner)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("app error: %v", err)
		os.Exit(1)
	}
}
