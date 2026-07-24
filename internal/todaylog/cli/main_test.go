package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCommandMapsConciseOptions(t *testing.T) {
	var got []string
	cmd := newCommand(func(input, debug, presets, history, prefilter string) error {
		got = []string{input, debug, presets, history, prefilter}
		return nil
	})

	err := cmd.Run(context.Background(), []string{
		"todaylog",
		"-d", "debug.log",
		"-p", `metric\.keep\.`,
		"--filter-presets", "presets.json",
		"--filter-history", "history.json",
		"today.log",
	})
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
	want := []string{"today.log", "debug.log", "presets.json", "history.json", `metric\.keep\.`}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("mapped options = %q want %q", got, want)
	}
}

func TestCommandKeepsHiddenInputCompatibility(t *testing.T) {
	var input string
	cmd := newCommand(func(path, _, _, _, _ string) error {
		input = path
		return nil
	})
	if err := cmd.Run(context.Background(), []string{"todaylog", "--input", "legacy.log"}); err != nil {
		t.Fatalf("run compatibility option: %v", err)
	}
	if input != "legacy.log" {
		t.Fatalf("input = %q want legacy.log", input)
	}
}

func TestCommandRejectsMissingAmbiguousAndExtraInput(t *testing.T) {
	for _, args := range [][]string{
		{"todaylog"},
		{"todaylog", "one.log", "--input", "two.log"},
		{"todaylog", "one.log", "two.log"},
	} {
		cmd := newCommand(func(_, _, _, _, _ string) error { return nil })
		if err := cmd.Run(context.Background(), args); err == nil {
			t.Fatalf("args %q unexpectedly succeeded", args)
		}
	}
}

func TestHelpShowsTodaylogOptionsOnce(t *testing.T) {
	var output bytes.Buffer
	cmd := newCommand(func(_, _, _, _, _ string) error { return nil })
	cmd.Writer = &output
	cmd.ErrWriter = &output
	if err := cmd.Run(context.Background(), []string{"todaylog", "--help"}); err != nil {
		t.Fatalf("show help: %v", err)
	}

	help := output.String()
	for _, want := range []string{
		"todaylog [options] FILE",
		"--debug FILE, -d FILE",
		"--prefilter REGEX, -p REGEX",
		"Configuration",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
	if strings.Contains(help, "--input") {
		t.Fatalf("help exposed compatibility input option:\n%s", help)
	}
	if count := strings.Count(help, "--version"); count != 1 {
		t.Fatalf("help contains --version %d times, want once:\n%s", count, help)
	}
}
