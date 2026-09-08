package mentions

import "strings"

func Extract(content string) []string {
	words := strings.Fields(content)

	mentions := make([]string, 0)
	seen := make(map[string]bool)

	for _, word := range words {
		if !strings.HasPrefix(word, "@") {
			continue
		}

		username := strings.TrimPrefix(word, "@")

		// Remove punctuation commonly placed after a mention.
		username = strings.TrimRight(
			username,
			".,!?;:)]}",
		)

		if username == "" {
			continue
		}

		username = strings.ToLower(username)

		if seen[username] {
			continue
		}

		seen[username] = true
		mentions = append(mentions, username)
	}

	return mentions
}
