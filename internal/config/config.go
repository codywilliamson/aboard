package config

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	TrelloAPIKey   string
	TrelloAPIToken string
	TrelloBoardID  string
	CodexCommand   []string
	ClaudeCommand  []string
}

// ConfigPath returns the path of the .env file that was loaded, or empty if none found.
var ConfigPath string

func Load(explicit string) Config {
	ConfigPath = loadConfig(explicit)

	return Config{
		TrelloAPIKey:   os.Getenv("TRELLO_API_KEY"),
		TrelloAPIToken: os.Getenv("TRELLO_API_TOKEN"),
		TrelloBoardID:  os.Getenv("TRELLO_BOARD_ID"),
		CodexCommand:   commandFromEnv("TRELLO_TUI_CODEX_COMMAND", []string{"codex"}),
		ClaudeCommand:  commandFromEnv("TRELLO_TUI_CLAUDE_COMMAND", []string{"claude"}),
	}
}

// loadConfig tries to load a .env file from the first location that exists.
// search order: explicit path > CWD/.env > <user config dir>/aboard/.env > next to executable
// returns the path that was loaded, or empty string.
func loadConfig(explicit string) string {
	if explicit != "" {
		if loadDotEnv(explicit) == nil {
			if _, err := os.Stat(explicit); err == nil {
				return explicit
			}
		}
		return ""
	}

	candidates := []string{".env"}

	if dir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(dir, "aboard", ".env"))
	}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			_ = loadDotEnv(path)
			abs, _ := filepath.Abs(path)
			return abs
		}
	}
	return ""
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
