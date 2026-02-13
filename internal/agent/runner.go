package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type AgentName string

const (
	AgentCodex  AgentName = "codex"
	AgentClaude AgentName = "claude"
)

type Runner struct {
	codexCmd  []string
	claudeCmd []string
	timeout   time.Duration
}

func NewRunner(codexCmd, claudeCmd []string) *Runner {
	return &Runner{
		codexCmd:  append([]string(nil), codexCmd...),
		claudeCmd: append([]string(nil), claudeCmd...),
		timeout:   90 * time.Second,
	}
}

func (r *Runner) Ask(ctx context.Context, agent AgentName, cardContext, userPrompt string) (string, error) {
	prompt := buildPrompt(cardContext, userPrompt)

	cmdSpec, err := r.commandFor(agent)
	if err != nil {
		return "", err
	}

	args := make([]string, 0, len(cmdSpec)-1)
	containsPromptPlaceholder := false
	for _, arg := range cmdSpec[1:] {
		replaced := strings.ReplaceAll(arg, "{prompt}", prompt)
		replaced = strings.ReplaceAll(replaced, "{context}", cardContext)
		if replaced != arg {
			containsPromptPlaceholder = true
		}
		args = append(args, replaced)
	}

	runCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, cmdSpec[0], args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if !containsPromptPlaceholder {
		cmd.Stdin = strings.NewReader(prompt)
	}

	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText != "" {
			return "", fmt.Errorf("agent command failed: %w: %s", err, errText)
		}
		return "", fmt.Errorf("agent command failed: %w", err)
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		if serr := strings.TrimSpace(stderr.String()); serr != "" {
			out = serr
		}
	}
	if out == "" {
		out = "(no output)"
	}

	return out, nil
}

func (r *Runner) commandFor(agent AgentName) ([]string, error) {
	switch agent {
	case AgentCodex:
		if len(r.codexCmd) == 0 {
			return nil, errors.New("codex command is not configured")
		}
		return r.codexCmd, nil
	case AgentClaude:
		if len(r.claudeCmd) == 0 {
			return nil, errors.New("claude command is not configured")
		}
		return r.claudeCmd, nil
	default:
		return nil, fmt.Errorf("unknown agent: %s", agent)
	}
}

func buildPrompt(cardContext, userPrompt string) string {
	return strings.TrimSpace(strings.Join([]string{
		"You are a Trello board assistant. You can perform actions by including <action>{...}</action> blocks in your response.",
		"",
		"Available actions:",
		`  <action>{"type":"move_card","card_id":"...","list_id":"..."}</action>`,
		`  <action>{"type":"update_card","card_id":"...","name":"...","desc":"..."}</action>`,
		`  <action>{"type":"add_comment","card_id":"...","text":"..."}</action>`,
		`  <action>{"type":"archive_card","card_id":"..."}</action>`,
		`  <action>{"type":"create_card","list_id":"...","name":"..."}</action>`,
		`  <action>{"type":"create_list","name":"..."}</action>`,
		`  <action>{"type":"archive_list","list_id":"..."}</action>`,
		"",
		"Board context:",
		cardContext,
		"",
		"User request:",
		userPrompt,
	}, "\n"))
}
