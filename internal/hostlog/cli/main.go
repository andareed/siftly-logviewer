package cli

import (
	"context"
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/hostlog"
	"github.com/urfave/cli/v3"
)

type runFunc func(inputPath, debugLogPath, filterPresetsPath, filterHistoryPath string) error

func Run(args []string) error {
	return newCommand(hostlog.Run).Run(context.Background(), args)
}

func newCommand(run runFunc) *cli.Command {
	return &cli.Command{
		Name:      "hostlog",
		Usage:     "Review host-log CSV and Siftly snapshots",
		UsageText: "hostlog [options] FILE",
		Version:   hostlog.Version,
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "file", UsageText: "FILE"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			inputPath, err := resolveInputPath(c)
			if err != nil {
				return err
			}
			return run(
				inputPath,
				c.String("debug"),
				c.String("filter-presets"),
				c.String("filter-history"),
			)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "debug",
				Aliases:   []string{"d"},
				Usage:     "Write diagnostic logs to `FILE`",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "input",
				Aliases:   []string{"i"},
				Usage:     "Compatibility alias for `FILE`",
				Hidden:    true,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "filter-presets",
				Usage:     "Read filter presets from `FILE` (default: hostlog-filters.json)",
				Category:  "Configuration",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "filter-history",
				Usage:     "Write filter history to `FILE` (default: hostlog-filter-history.json)",
				Category:  "Configuration",
				TakesFile: true,
			},
		},
	}
}

func resolveInputPath(c *cli.Command) (string, error) {
	if c.Args().Present() {
		return "", fmt.Errorf("unexpected argument %q", c.Args().First())
	}
	positional := c.StringArg("file")
	compatibilityFlag := c.String("input")
	if positional != "" && compatibilityFlag != "" {
		return "", fmt.Errorf("use FILE or --input, not both")
	}
	if compatibilityFlag != "" {
		return compatibilityFlag, nil
	}
	if positional == "" {
		return "", fmt.Errorf("missing FILE")
	}
	return positional, nil
}
