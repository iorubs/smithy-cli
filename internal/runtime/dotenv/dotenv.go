// Package dotenv loads simple KEY=VALUE files in the style of
// docker-compose's `.env`. Values already present in the process
// environment are preserved (the shell wins), matching compose
// semantics. No variable expansion, no `export` prefix; quoted
// values strip surrounding single or double quotes.
package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Load reads path, parses KEY=VALUE pairs, and applies them to the
// process environment via os.Setenv. Existing env keys are not
// overwritten. Missing files are not an error: Load returns nil so
// callers can treat the file as optional.
func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("dotenv: open %s: %w", path, err)
	}
	defer f.Close()

	pairs, err := parse(f)
	if err != nil {
		return fmt.Errorf("dotenv: parse %s: %w", path, err)
	}
	for k, v := range pairs {
		if _, present := os.LookupEnv(k); present {
			continue
		}
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("dotenv: setenv %s: %w", k, err)
		}
	}
	return nil
}

func parse(r io.Reader) (map[string]string, error) {
	out := map[string]string{}
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			return nil, fmt.Errorf("invalid line: %q", line)
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if i := indexUnquotedHash(val); i >= 0 {
			val = strings.TrimSpace(val[:i])
		}
		val = stripQuotes(val)
		out[key] = val
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func indexUnquotedHash(s string) int {
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote == 0 && (c == '"' || c == '\''):
			quote = c
		case quote != 0 && c == quote:
			quote = 0
		case quote == 0 && c == '#':
			return i
		}
	}
	return -1
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
