package cli

import (
	"context"
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/todaylog"
	"github.com/urfave/cli/v3"
)

type runFunc func(inputPath, debugLogPath, filterPresetsPath, filterHistoryPath, prefilter string) error

func Run(args []string) error {
	return newCommand(todaylog.Run).Run(context.Background(), args)
}

func newCommand(run runFunc) *cli.Command {
	return &cli.Command{
		Name:      "todaylog",
		Usage:     "Review raw today logs and Siftly snapshots",
		UsageText: "todaylog [options] FILE",
		Version:   todaylog.Version,
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
				c.String("prefilter"),
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
				Usage:     "Read filter presets from `FILE` (default: todaylog-filters.json)",
				Category:  "Configuration",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "filter-history",
				Usage:     "Write filter history to `FILE` (default: todaylog-filter-history.json)",
				Category:  "Configuration",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:    "prefilter",
				Aliases: []string{"p"},
				Usage:   "Load only raw lines matching `REGEX`",
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
