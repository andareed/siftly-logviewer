package siftly

import (
	"fmt"
	"strconv"
	"strings"
)

// toggleColumnsBySpec flips visibility for the columns referenced in spec.
// spec accepts comma or space separated column names or 1-based indices.
// Returns the list of column display names toggled and any tokens that did not match.
func (m *Model) toggleColumnsBySpec(spec string) (toggled []string, missing []string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil, fmt.Errorf("no columns specified")
	}

	tokens, err := parseColumnTokens(spec)
	if err != nil {
		return nil, nil, err
	}
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("no columns specified")
	}

	seen := make(map[int]struct{})
	for _, token := range tokens {
		if token == "" {
			continue
		}
		idx, ok := m.resolveColumnIndex(token)
		if !ok {
			missing = append(missing, token)
			continue
		}
		if _, dup := seen[idx]; dup {
			continue
		}
		m.table.header[idx].Visible = !m.table.header[idx].Visible
		toggled = append(toggled, m.table.header[idx].Name)
		seen[idx] = struct{}{}
	}

	m.refreshView("toggle-columns", true)
	return toggled, missing, nil
}

// resolveColumnIndex resolves a user provided token to a column index.
// Supports 1-based indices or case-insensitive exact name matches.
func (m *Model) resolveColumnIndex(token string) (int, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return -1, false
	}

	if n, err := strconv.Atoi(token); err == nil {
		n-- // convert to 0-based
		if n >= 0 && n < len(m.table.header) {
			return n, true
		}
	}

	token = strings.ToLower(token)
	for i, col := range m.table.header {
		if strings.ToLower(col.Name) == token {
			return i, true
		}
	}
	return -1, false
}

// parseColumnTokens splits a user column list into tokens supporting quotes.
// Delimiters are comma or whitespace when not inside quotes.
func parseColumnTokens(spec string) ([]string, error) {
	var tokens []string
	var b strings.Builder
	inQuote := false
	quoteChar := byte(0)

	flush := func() {
		if b.Len() > 0 {
			tokens = append(tokens, b.String())
			b.Reset()
		}
	}

	for i := 0; i < len(spec); i++ {
		c := spec[i]
		switch {
		case inQuote:
			if c == quoteChar {
				inQuote = false
				continue
			}
			if c == '\\' && i+1 < len(spec) && spec[i+1] == quoteChar {
				b.WriteByte(quoteChar)
				i++
				continue
			}
			b.WriteByte(c)
		case c == '"' || c == '\'':
			inQuote = true
			quoteChar = c
		case c == ',' || c == ' ' || c == '\t' || c == '\n' || c == '\r':
			flush()
		default:
			b.WriteByte(c)
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated quote")
	}
	flush()
	return tokens, nil
}
