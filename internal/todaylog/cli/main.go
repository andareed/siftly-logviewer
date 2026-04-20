package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/andareed/siftly-hostlog/internal/todaylog"
	"github.com/urfave/cli/v3"
)

func Run() {
	app := &cli.Command{
		Name:    "todaylog",
		Usage:   "Siftly Toolkit: Today Log Viewer",
		Version: todaylog.Version,
		Action: func(_ context.Context, c *cli.Command) error {
			if c.Bool("version") {
				fmt.Println("Version:", todaylog.Version)
				return nil
			}

			inputPath := c.String("input")
			if inputPath == "" {
				inputPath = c.Args().First()
			}

			if inputPath == "" {
				_ = cli.ShowAppHelp(c)
				return cli.Exit("", 1)
			}

			return todaylog.Run(
				inputPath,
				c.String("debug"),
				c.String("filter-presets"),
				c.String("filter-history"),
			)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "debug",
				Usage: "Write debug logs to file",
			},
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Path to input file (.csv or .json)",
			},
			&cli.StringFlag{
				Name:  "filter-presets",
				Usage: "Path to app-specific filter presets JSON (default: todaylog-filters.json)",
			},
			&cli.StringFlag{
				Name:  "filter-history",
				Usage: "Path to writable filter history JSON (default: todaylog-filter-history.json)",
			},
			&cli.BoolFlag{
				Name:  "version",
				Usage: "Print version and exit",
			},
		},
	}

	_ = app.Run(context.Background(), os.Args)
}
