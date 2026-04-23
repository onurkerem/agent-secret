package prompt

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptSecret prompts the user for a secret value with hidden input
func PromptSecret(promptText string) (string, error) {
	fmt.Fprint(os.Stderr, promptText)

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	var input strings.Builder
	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		if n == 0 {
			continue
		}

		switch buf[0] {
		case '\r', '\n':
			// Enter pressed
			fmt.Fprintln(os.Stderr)
			return input.String(), nil
		case 3: // Ctrl+C
			fmt.Fprintln(os.Stderr, "^C")
			os.Exit(1)
		case 127, 8: // Backspace
			if input.Len() > 0 {
				// Remove last character
				str := input.String()
				input.Reset()
				input.WriteString(str[:len(str)-1])
				// Move cursor back, clear character, move cursor back again
				fmt.Fprint(os.Stderr, "\b \b")
			}
		default:
			// Only accept printable characters
			if buf[0] >= 32 && buf[0] < 127 {
				input.WriteByte(buf[0])
				fmt.Fprint(os.Stderr, "*")
			}
		}
	}
}

// PromptConfirm prompts the user for a yes/no confirmation
func PromptConfirm(promptText string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", promptText)

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return false, fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	buf := make([]byte, 1)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	fmt.Fprintln(os.Stderr)

	if n == 0 {
		return false, nil
	}

	response := strings.ToLower(string(buf[0]))
	return response == "y", nil
}
