package util

import (
	"bufio"
	"strings"
)

func RemoveCommentsAndEmptyLines(prefix, input string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip if the entire line is a comment
		if strings.HasPrefix(trimmedLine, prefix) {
			continue
		}

		// Process comments in the line
		if idx := strings.Index(line, prefix); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Add only if it is not an empty line
		if line != "" {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
