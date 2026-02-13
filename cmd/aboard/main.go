package main

import (
	"flag"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/codywilliamson/aboard/internal/agent"
	"github.com/codywilliamson/aboard/internal/config"
	"github.com/codywilliamson/aboard/internal/trello"
	"github.com/codywilliamson/aboard/internal/ui"
)

func main() {
	configPath := flag.String("config", "", "path to .env config file")
	flag.StringVar(configPath, "c", "", "path to .env config file (shorthand)")
	flag.Parse()

	cfg := config.Load(*configPath)

	trelloClient := trello.NewClient(cfg.TrelloAPIKey, cfg.TrelloAPIToken)
	runner := agent.NewRunner(cfg.CodexCommand, cfg.ClaudeCommand)

	m := ui.NewModel(cfg, trelloClient, runner)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("app error: %v", err)
		os.Exit(1)
	}
}
