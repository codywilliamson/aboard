package config

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type Config struct {
	TrelloAPIKey   string
	TrelloAPIToken string
	TrelloBoardID  string
	CodexCommand   []string
	ClaudeCommand  []string
}

func LoadFromEnv() Config {
	_ = loadDotEnv(".env")

	return Config{
		TrelloAPIKey:   os.Getenv("TRELLO_API_KEY"),
		TrelloAPIToken: os.Getenv("TRELLO_API_TOKEN"),
		TrelloBoardID:  os.Getenv("TRELLO_BOARD_ID"),
		CodexCommand:   commandFromEnv("TRELLO_TUI_CODEX_COMMAND", []string{"codex"}),
		ClaudeCommand:  commandFromEnv("TRELLO_TUI_CLAUDE_COMMAND", []string{"claude"}),
	}
}

func commandFromEnv(name string, fallback []string) []string {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}

	var cmd []string
	if err := json.Unmarshal([]byte(raw), &cmd); err != nil {
		return fallback
	}
	if len(cmd) == 0 {
		return fallback
	}
	return cmd
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(parts[1])

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}
