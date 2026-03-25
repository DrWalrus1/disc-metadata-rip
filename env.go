package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadEnv reads a .env file and sets any unset environment variables.
// Existing environment variables are not overwritten.
func loadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env is optional
		}
		return fmt.Errorf("opening .env: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return fmt.Errorf(".env line %d: missing '=' in %q", lineNum, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Strip optional surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Don't overwrite existing env vars
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
