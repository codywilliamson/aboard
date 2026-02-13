package ui

import "strings"

func ellipsis(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}

func shortID(s string) string {
	if len(s) <= 8 {
		return s
	}
	return s[:8]
}

func wrapForPane(s string, width int) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			out = append(out, "")
			continue
		}
		for len(line) > width {
			cut := strings.LastIndex(line[:width], " ")
			if cut <= 0 {
				cut = width
			}
			out = append(out, strings.TrimSpace(line[:cut]))
			line = strings.TrimSpace(line[cut:])
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func stripMarkdown(s string) string {
	// strip bold/italic markers
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	// strip markdown link syntax [text](url) → text (url)
	for {
		open := strings.Index(s, "[")
		if open < 0 {
			break
		}
		close := strings.Index(s[open:], "](")
		if close < 0 {
			break
		}
		close += open
		end := strings.Index(s[close:], ")")
		if end < 0 {
			break
		}
		end += close
		text := s[open+1 : close]
		url := s[close+2 : end]
		s = s[:open] + text + " (" + url + ")" + s[end+1:]
	}
	return s
}

func clipToLineCount(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}
