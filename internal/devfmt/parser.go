package devfmt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type ParsedRow struct {
	Category     string
	ID           string
	Time         string
	FieldName    string
	FieldValue   string
	FieldValueLC string
	AgentID      string
	Extra        string
}

func ParseDump(r io.Reader, onRow func(ParsedRow) error) error {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 8*1024*1024)

	inCopy := false
	inInsert := false
	var insertBuf strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if inCopy {
			if trimmed == "\\." {
				inCopy = false
				continue
			}
			row, ok, err := parseCopyLine(line)
			if err != nil {
				return err
			}
			if ok {
				if err := onRow(row); err != nil {
					return err
				}
			}
			continue
		}

		if inInsert {
			insertBuf.WriteString("\n")
			insertBuf.WriteString(line)
			if strings.Contains(line, ";") {
				if err := parseInsertStatement(insertBuf.String(), onRow); err != nil {
					return err
				}
				insertBuf.Reset()
				inInsert = false
			}
			continue
		}

		if isCopyStart(trimmed) {
			inCopy = true
			continue
		}

		if isInsertStart(trimmed) {
			insertBuf.WriteString(line)
			if strings.Contains(line, ";") {
				if err := parseInsertStatement(insertBuf.String(), onRow); err != nil {
					return err
				}
				insertBuf.Reset()
			} else {
				inInsert = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if inInsert {
		return fmt.Errorf("unterminated INSERT statement")
	}
	return nil
}

func isCopyStart(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	if !strings.HasPrefix(normalized, "copy public.devinfo") {
		return false
	}
	return strings.Contains(normalized, "from stdin")
}

func isInsertStart(line string) bool {
	normalized := strings.ToLower(strings.TrimSpace(line))
	return strings.HasPrefix(normalized, "insert into public.devinfo")
}

func parseCopyLine(line string) (ParsedRow, bool, error) {
	parts := strings.Split(line, "\t")
	if len(parts) != 8 {
		return ParsedRow{}, false, nil
	}
	for i := range parts {
		parts[i] = decodeCopyValue(parts[i])
	}
	return ParsedRow{
		Category:     parts[0],
		ID:           parts[1],
		Time:         parts[2],
		FieldName:    parts[3],
		FieldValue:   parts[4],
		FieldValueLC: parts[5],
		AgentID:      parts[6],
		Extra:        parts[7],
	}, true, nil
}

func decodeCopyValue(in string) string {
	if in == `\\N` {
		return ""
	}
	if !strings.Contains(in, `\\`) {
		return in
	}
	repl := strings.NewReplacer(
		`\\t`, "\t",
		`\\n`, "\n",
		`\\r`, "\r",
		`\\\\`, `\\`,
	)
	return repl.Replace(in)
}

func parseInsertStatement(stmt string, onRow func(ParsedRow) error) error {
	lower := strings.ToLower(stmt)
	idx := strings.Index(lower, "values")
	if idx < 0 {
		return nil
	}
	values := strings.TrimSpace(stmt[idx+len("values"):])
	if strings.HasSuffix(values, ";") {
		values = strings.TrimSuffix(values, ";")
	}
	tuples, err := parseSQLTuples(values)
	if err != nil {
		return err
	}
	for _, t := range tuples {
		if len(t) != 8 {
			continue
		}
		row := ParsedRow{
			Category:     sqlValue(t[0]),
			ID:           sqlValue(t[1]),
			Time:         sqlValue(t[2]),
			FieldName:    sqlValue(t[3]),
			FieldValue:   sqlValue(t[4]),
			FieldValueLC: sqlValue(t[5]),
			AgentID:      sqlValue(t[6]),
			Extra:        sqlValue(t[7]),
		}
		if err := onRow(row); err != nil {
			return err
		}
	}
	return nil
}

func parseSQLTuples(values string) ([][]string, error) {
	var tuples [][]string
	var currentTuple []string
	var token strings.Builder
	inString := false
	depth := 0

	for i := 0; i < len(values); i++ {
		ch := values[i]

		if inString {
			if ch == '\'' {
				if i+1 < len(values) && values[i+1] == '\'' {
					token.WriteByte('\'')
					i++
					continue
				}
				inString = false
				continue
			}
			token.WriteByte(ch)
			continue
		}

		switch ch {
		case '\'':
			inString = true
		case '(':
			if depth == 0 {
				currentTuple = make([]string, 0, 8)
				token.Reset()
			}
			depth++
		case ')':
			if depth == 1 {
				currentTuple = append(currentTuple, strings.TrimSpace(token.String()))
				token.Reset()
				tuples = append(tuples, currentTuple)
				currentTuple = nil
			}
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 1 {
				currentTuple = append(currentTuple, strings.TrimSpace(token.String()))
				token.Reset()
			}
		default:
			if depth > 0 {
				token.WriteByte(ch)
			}
		}
	}

	if inString || depth != 0 {
		return nil, fmt.Errorf("malformed INSERT VALUES")
	}
	return tuples, nil
}

func sqlValue(tok string) string {
	t := strings.TrimSpace(tok)
	if strings.EqualFold(t, "null") {
		return ""
	}
	return t
}
