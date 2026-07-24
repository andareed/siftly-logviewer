package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCommandMapsConciseOptions(t *testing.T) {
	var got []string
	cmd := newCommand(func(input, debug, presets, history string) error {
		got = []string{input, debug, presets, history}
		return nil
	})

	err := cmd.Run(context.Background(), []string{
		"hostlog",
		"-d", "debug.log",
		"--filter-presets", "presets.json",
		"--filter-history", "history.json",
		"host.csv",
	})
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
	want := []string{"host.csv", "debug.log", "presets.json", "history.json"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("mapped options = %q want %q", got, want)
	}
}

func TestCommandKeepsHiddenInputCompatibility(t *testing.T) {
	var input string
	cmd := newCommand(func(path, _, _, _ string) error {
		input = path
		return nil
	})
	if err := cmd.Run(context.Background(), []string{"hostlog", "-i", "legacy.csv"}); err != nil {
		t.Fatalf("run compatibility option: %v", err)
	}
	if input != "legacy.csv" {
		t.Fatalf("input = %q want legacy.csv", input)
	}
}

func TestCommandRejectsMissingAmbiguousAndExtraInput(t *testing.T) {
	for _, args := range [][]string{
		{"hostlog"},
		{"hostlog", "one.csv", "--input", "two.csv"},
		{"hostlog", "one.csv", "two.csv"},
	} {
		cmd := newCommand(func(_, _, _, _ string) error { return nil })
		if err := cmd.Run(context.Background(), args); err == nil {
			t.Fatalf("args %q unexpectedly succeeded", args)
		}
	}
}

func TestCommandPropagatesViewerError(t *testing.T) {
	want := errors.New("viewer failed")
	cmd := newCommand(func(_, _, _, _ string) error { return want })
	if err := cmd.Run(context.Background(), []string{"hostlog", "host.csv"}); !errors.Is(err, want) {
		t.Fatalf("error = %v want %v", err, want)
	}
}

func TestHelpShowsCanonicalInterface(t *testing.T) {
	var output bytes.Buffer
	cmd := newCommand(func(_, _, _, _ string) error { return nil })
	cmd.Writer = &output
	cmd.ErrWriter = &output
	if err := cmd.Run(context.Background(), []string{"hostlog", "--help"}); err != nil {
		t.Fatalf("show help: %v", err)
	}

	help := output.String()
	for _, want := range []string{"hostlog [options] FILE", "--debug FILE, -d FILE", "Configuration"} {
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
