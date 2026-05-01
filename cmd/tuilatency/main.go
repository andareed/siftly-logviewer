package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type config struct {
	mode      string
	fps       int
	bodyLines int
	bodyWidth int
	batchMS   int
	altScreen bool
	logPath   string
}

type metricLogger struct {
	w      io.WriteCloser
	writer *bufio.Writer
}

func main() {
	cfg := parseFlags()

	logger, err := openMetricLogger(cfg.logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	switch cfg.mode {
	case "raw":
		err = runRawProbe(logger)
	case "raw-silent":
		err = runRawSilentProbe(logger)
	case "raw-inline":
		err = runRawInlineProbe(cfg, logger)
	case "raw-batch":
		err = runRawBatchProbe(cfg, logger)
	case "raw-footer":
		err = runRawFrameProbe(cfg, logger, false)
	case "raw-full":
		err = runRawFrameProbe(cfg, logger, true)
	case "tea-min":
		err = runTeaProbe(cfg, logger, false)
	case "tea-full":
		err = runTeaProbe(cfg, logger, true)
	default:
		err = fmt.Errorf("unknown mode %q", cfg.mode)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	cfg := config{}
	flag.StringVar(&cfg.mode, "mode", "tea-min", "probe mode: raw, raw-silent, raw-inline, raw-batch, raw-footer, raw-full, tea-min, tea-full")
	flag.IntVar(&cfg.fps, "fps", 60, "Bubble Tea renderer FPS for tea-* modes")
	flag.IntVar(&cfg.bodyLines, "body-lines", 40, "static body line count for tea-full")
	flag.IntVar(&cfg.bodyWidth, "body-width", 160, "static body width for tea-full")
	flag.IntVar(&cfg.batchMS, "batch-ms", 250, "redraw interval for raw-batch")
	flag.BoolVar(&cfg.altScreen, "alt-screen", true, "use Bubble Tea alt screen for tea-* modes")
	flag.StringVar(&cfg.logPath, "log", "", "optional metrics log path")
	flag.Parse()
	return cfg
}

func openMetricLogger(path string) (*metricLogger, error) {
	if strings.TrimSpace(path) == "" {
		return &metricLogger{}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &metricLogger{w: f, writer: bufio.NewWriter(f)}, nil
}

func (l *metricLogger) Printf(format string, args ...any) {
	if l == nil || l.writer == nil {
		return
	}
	fmt.Fprintf(l.writer, format, args...)
	l.writer.Flush()
}

func (l *metricLogger) Close() error {
	if l == nil || l.w == nil {
		return nil
	}
	if l.writer != nil {
		l.writer.Flush()
	}
	return l.w.Close()
}

func runRawProbe(logger *metricLogger) error {
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(fd, oldState)

	fmt.Fprint(os.Stdout, "\x1b[2J\x1b[Hraw terminal probe\r\n")
	fmt.Fprint(os.Stdout, "type to measure; ctrl-c or esc exits\r\n\r\n")

	buf := make([]byte, 16)
	started := time.Now()
	last := started
	count := 0

	for {
		n, err := os.Stdin.Read(buf)
		now := time.Now()
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			if b == 3 || b == 27 {
				fmt.Fprint(os.Stdout, "\r\nexit\r\n")
				return nil
			}

			count++
			sinceLast := now.Sub(last)
			last = now
			writeStart := time.Now()
			fmt.Fprintf(os.Stdout, "\rkey=%q count=%d since_last=%s uptime=%s read_n=%d read_idx=%d\x1b[K", printableByte(b), count, sinceLast.Round(time.Millisecond), now.Sub(started).Round(time.Millisecond), n, i)
			writeDur := time.Since(writeStart)
			logger.Printf("mode=raw count=%d byte=%d since_last_us=%d write_us=%d read_n=%d read_idx=%d\n", count, b, sinceLast.Microseconds(), writeDur.Microseconds(), n, i)
		}
	}
}

func runRawFrameProbe(cfg config, logger *metricLogger, full bool) error {
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(fd, oldState)

	body := buildStaticBody(cfg.bodyLines, cfg.bodyWidth)
	status := padRight("type to measure; ctrl-c or esc exits", cfg.bodyWidth)
	fmt.Fprint(os.Stdout, "\x1b[?1049h\x1b[2J\x1b[H"+body+"\n"+status)
	defer fmt.Fprint(os.Stdout, "\x1b[?1049l")

	buf := make([]byte, 16)
	started := time.Now()
	last := started
	count := 0
	input := ""
	mode := "raw-footer"
	if full {
		mode = "raw-full"
	}

	for {
		n, err := os.Stdin.Read(buf)
		now := time.Now()
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			if b == 3 || b == 27 {
				return nil
			}

			if b == 127 || b == 8 {
				input = trimLastRune(input)
			} else if b >= 32 && b < 127 {
				input += string(b)
			}

			count++
			sinceLast := now.Sub(last)
			last = now
			status = padRight(fmt.Sprintf(
				"mode=%s count=%d len=%d since_last=%s input=%q read_n=%d read_idx=%d",
				mode,
				count,
				len(input),
				sinceLast.Round(time.Millisecond),
				tail(input, 50),
				n,
				i,
			), cfg.bodyWidth)

			writeStart := time.Now()
			if full {
				fmt.Fprint(os.Stdout, "\x1b[H"+body+"\n"+status+"\x1b[K")
			} else {
				fmt.Fprint(os.Stdout, "\x1b[H"+strings.Repeat("\n", cfg.bodyLines)+status+"\x1b[K")
			}
			writeDur := time.Since(writeStart)
			logger.Printf("mode=%s count=%d byte=%d since_last_us=%d write_us=%d read_n=%d read_idx=%d\n", mode, count, b, sinceLast.Microseconds(), writeDur.Microseconds(), n, i)
		}
	}
}

func runRawSilentProbe(logger *metricLogger) error {
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(fd, oldState)

	fmt.Fprint(os.Stdout, "raw silent probe: type normally; ctrl-c or esc exits\r\n")

	buf := make([]byte, 16)
	started := time.Now()
	last := started
	count := 0

	for {
		n, err := os.Stdin.Read(buf)
		now := time.Now()
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			if b == 3 || b == 27 {
				fmt.Fprintf(os.Stdout, "\r\nexit count=%d uptime=%s\r\n", count, now.Sub(started).Round(time.Millisecond))
				return nil
			}

			count++
			sinceLast := now.Sub(last)
			last = now
			logger.Printf("mode=raw-silent count=%d byte=%d since_last_us=%d read_n=%d read_idx=%d\n", count, b, sinceLast.Microseconds(), n, i)
		}
	}
}

func runRawInlineProbe(cfg config, logger *metricLogger) error {
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(fd, oldState)

	fmt.Fprint(os.Stdout, "\x1b[2J\x1b[Hraw inline probe: type to measure; ctrl-c or esc exits\r\n\r\n")

	buf := make([]byte, 16)
	started := time.Now()
	last := started
	count := 0
	input := ""

	for {
		n, err := os.Stdin.Read(buf)
		now := time.Now()
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			if b == 3 || b == 27 {
				fmt.Fprint(os.Stdout, "\r\nexit\r\n")
				return nil
			}

			if b == 127 || b == 8 {
				input = trimLastRune(input)
			} else if b >= 32 && b < 127 {
				input += string(b)
			}

			count++
			sinceLast := now.Sub(last)
			last = now
			line := fmt.Sprintf(
				"mode=raw-inline count=%d len=%d since_last=%s input=%q read_n=%d read_idx=%d",
				count,
				len(input),
				sinceLast.Round(time.Millisecond),
				tail(input, max(20, cfg.bodyWidth/2)),
				n,
				i,
			)

			writeStart := time.Now()
			fmt.Fprint(os.Stdout, "\r"+line+"\x1b[K")
			writeDur := time.Since(writeStart)
			logger.Printf("mode=raw-inline count=%d byte=%d since_last_us=%d write_us=%d read_n=%d read_idx=%d\n", count, b, sinceLast.Microseconds(), writeDur.Microseconds(), n, i)
		}
	}
}

func runRawBatchProbe(cfg config, logger *metricLogger) error {
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(fd, oldState)

	if cfg.batchMS < 1 {
		cfg.batchMS = 250
	}
	interval := time.Duration(cfg.batchMS) * time.Millisecond
	fmt.Fprintf(os.Stdout, "\x1b[2J\x1b[Hraw batch probe: redraw every %s; ctrl-c or esc exits\r\n\r\n", interval)

	buf := make([]byte, 16)
	started := time.Now()
	last := started
	lastDraw := started
	count := 0
	input := ""

	for {
		n, err := os.Stdin.Read(buf)
		now := time.Now()
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		if n == 0 {
			continue
		}

		for i := 0; i < n; i++ {
			b := buf[i]
			if b == 3 || b == 27 {
				fmt.Fprint(os.Stdout, "\r\nexit\r\n")
				return nil
			}

			if b == 127 || b == 8 {
				input = trimLastRune(input)
			} else if b >= 32 && b < 127 {
				input += string(b)
			}

			count++
			sinceLast := now.Sub(last)
			last = now
			drew := false
			writeDur := time.Duration(0)

			if now.Sub(lastDraw) >= interval {
				line := fmt.Sprintf(
					"mode=raw-batch count=%d len=%d since_last=%s input=%q read_n=%d read_idx=%d",
					count,
					len(input),
					sinceLast.Round(time.Millisecond),
					tail(input, max(20, cfg.bodyWidth/2)),
					n,
					i,
				)
				writeStart := time.Now()
				fmt.Fprint(os.Stdout, "\r"+line+"\x1b[K")
				writeDur = time.Since(writeStart)
				lastDraw = now
				drew = true
			}

			logger.Printf("mode=raw-batch count=%d byte=%d since_last_us=%d drew=%t write_us=%d read_n=%d read_idx=%d\n", count, b, sinceLast.Microseconds(), drew, writeDur.Microseconds(), n, i)
		}
	}
}

func printableByte(b byte) string {
	if b >= 32 && b < 127 {
		return string(b)
	}
	return fmt.Sprintf("\\x%02x", b)
}

type probeModel struct {
	full     bool
	body     string
	width    int
	input    string
	count    int
	started  time.Time
	lastKey  time.Time
	keyTime  time.Time
	updateUS int64
	viewUS   int64
	logger   *metricLogger
}

func newProbeModel(cfg config, logger *metricLogger, full bool) *probeModel {
	return &probeModel{
		full:    full,
		body:    buildStaticBody(cfg.bodyLines, cfg.bodyWidth),
		width:   cfg.bodyWidth,
		started: time.Now(),
		logger:  logger,
	}
}

func (m *probeModel) Init() tea.Cmd { return nil }

func (m *probeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	start := time.Now()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		now := time.Now()
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyBackspace:
			m.input = trimLastRune(m.input)
		default:
			if len(msg.Runes) > 0 {
				m.input += string(msg.Runes)
			}
		}
		m.count++
		m.lastKey = m.keyTime
		m.keyTime = now
		m.updateUS = time.Since(start).Microseconds()
	}
	return m, nil
}

func (m *probeModel) View() string {
	start := time.Now()
	defer func() { m.viewUS = time.Since(start).Microseconds() }()

	line := m.statusLine()
	if !m.full {
		return line + "\n\nesc/ctrl-c exits"
	}
	return m.body + "\n" + line
}

func (m *probeModel) statusLine() string {
	now := time.Now()
	sinceLast := time.Duration(0)
	if !m.lastKey.IsZero() && !m.keyTime.IsZero() {
		sinceLast = m.keyTime.Sub(m.lastKey)
	}
	keyToView := time.Duration(0)
	if !m.keyTime.IsZero() {
		keyToView = now.Sub(m.keyTime)
	}

	m.logger.Printf(
		"mode=tea full=%t count=%d input_len=%d since_last_us=%d key_to_view_us=%d update_us=%d prev_view_us=%d\n",
		m.full,
		m.count,
		len(m.input),
		sinceLast.Microseconds(),
		keyToView.Microseconds(),
		m.updateUS,
		m.viewUS,
	)

	status := fmt.Sprintf(
		"count=%d len=%d since_last=%s key_to_view=%s update=%dus prev_view=%dus input=%q",
		m.count,
		len(m.input),
		sinceLast.Round(time.Millisecond),
		keyToView.Round(time.Millisecond),
		m.updateUS,
		m.viewUS,
		tail(m.input, 50),
	)
	if m.width > 0 && lipgloss.Width(status) < m.width {
		status += strings.Repeat(" ", m.width-lipgloss.Width(status))
	}
	return status
}

func runTeaProbe(cfg config, logger *metricLogger, full bool) error {
	model := newProbeModel(cfg, logger, full)
	opts := []tea.ProgramOption{tea.WithFPS(cfg.fps)}
	if cfg.altScreen {
		opts = append(opts, tea.WithAltScreen())
	}
	_, err := tea.NewProgram(model, opts...).Run()
	return err
}

func buildStaticBody(lines int, width int) string {
	if lines < 1 {
		lines = 1
	}
	if width < 20 {
		width = 20
	}
	row := strings.Repeat(".", width)
	out := make([]string, 0, lines)
	for i := 0; i < lines; i++ {
		label := fmt.Sprintf("%03d ", i+1)
		out = append(out, label+row[len(label):])
	}
	return strings.Join(out, "\n")
}

func trimLastRune(s string) string {
	if s == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

func tail(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[len(r)-maxRunes:])
}

func padRight(s string, width int) string {
	if width <= 0 {
		return s
	}
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
