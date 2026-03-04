package pluginlog

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	recordStartRE = regexp.MustCompile(`(?m)^sw(?:-(?:[1-9]|1[0-9]|20))?:\d+:\d+(?:\.\d+)?:`)
	recordHeadRE  = regexp.MustCompile(`(?s)^(sw(?:-(?:[1-9]|1[0-9]|20))?):(\d+):(\d+(?:\.\d+)?):(.*)$`)
	localTimeRE   = regexp.MustCompile(`(?s)^([A-Z][a-z]{2} [A-Z][a-z]{2}\s+\d{1,2} \d{2}:\d{2}:\d{2} \d{4}):\s*(.*)$`)
	levelPrefixRE = regexp.MustCompile(`^([A-Za-z]+):\s+`)
	mflRE         = regexp.MustCompile(`(?s)^([^:\n]+)::([^:\n]+):(\d+):\s*(.*)$`)
)

const (
	logLocalTimeLayout = "Mon Jan 2 15:04:05 2006"
	tableTimeLayout    = "2006-01-02 15:04:05"
)

func parsePluginLog(content string) ([][]string, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	starts := recordStartRE.FindAllStringIndex(normalized, -1)
	if len(starts) == 0 {
		return nil, fmt.Errorf("no plugin log records found")
	}

	records := [][]string{pluginLogHeader()}
	for i, start := range starts {
		begin := start[0]
		end := len(normalized)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}

		rawRecord := strings.Trim(normalized[begin:end], "\n")
		if strings.TrimSpace(rawRecord) == "" {
			continue
		}

		row, err := parsePluginRecord(rawRecord)
		if err != nil {
			return nil, fmt.Errorf("parse record %d: %w", i+1, err)
		}
		records = append(records, row)
	}

	if len(records) == 1 {
		return nil, fmt.Errorf("no parseable plugin log records found")
	}

	return records, nil
}

func pluginLogHeader() []string {
	return []string{"process", "pid", "epoch", "time", "level", "module", "function", "line", "message"}
}

func parsePluginRecord(raw string) ([]string, error) {
	head := recordHeadRE.FindStringSubmatch(raw)
	if head == nil {
		return nil, fmt.Errorf("invalid plugin record format")
	}

	process := strings.TrimSpace(head[1])
	pid := strings.TrimSpace(head[2])
	epoch := strings.TrimSpace(head[3])
	tail := strings.TrimSpace(head[4])

	localRaw := ""
	payload := tail
	if tm := localTimeRE.FindStringSubmatch(tail); tm != nil {
		localRaw = strings.TrimSpace(tm[1])
		payload = strings.TrimSpace(tm[2])
	}

	timeValue := normalizeTableTime(localRaw)
	level, module, function, codeLine, message := parsePayload(payload)

	return []string{
		process,
		pid,
		epoch,
		timeValue,
		level,
		module,
		function,
		codeLine,
		message,
	}, nil
}

func normalizeTableTime(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	t, err := time.Parse(logLocalTimeLayout, raw)
	if err != nil {
		return raw
	}
	return t.Format(tableTimeLayout)
}

func parsePayload(payload string) (level, module, function, codeLine, message string) {
	working := strings.TrimSpace(payload)
	if working == "" {
		return "", "", "", "", ""
	}

	original := working
	if m := levelPrefixRE.FindStringSubmatch(working); m != nil {
		level = m[1]
		working = strings.TrimSpace(working[len(m[0]):])
	}

	if m := mflRE.FindStringSubmatch(working); m != nil {
		module = strings.TrimSpace(m[1])
		function = strings.TrimSpace(m[2])
		codeLine = strings.TrimSpace(m[3])
		message = strings.TrimSpace(m[4])
		if level == "" {
			if lm := levelPrefixRE.FindStringSubmatch(message); lm != nil {
				level = lm[1]
				message = strings.TrimSpace(message[len(lm[0]):])
			}
		}
		return
	}

	if level != "" {
		message = working
	} else {
		message = original
	}
	return
}
