package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/andareed/siftly-hostlog/internal/pluginlog"
	"github.com/urfave/cli/v3"
)

func Run() {
	app := &cli.Command{
		Name:    "pluginlog",
		Usage:   "Siftly Toolkit: Plugin Log Viewer",
		Version: pluginlog.Version,
		Action: func(_ context.Context, c *cli.Command) error {
			if c.Bool("version") {
				fmt.Println("Version:", pluginlog.Version)
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

			return pluginlog.Run(inputPath, c.String("debug"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "debug",
				Usage: "Write debug logs to file",
			},
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Path to input file (.log or .json)",
			},
			&cli.BoolFlag{
				Name:  "version",
				Usage: "Print version and exit",
			},
		},
	}

	_ = app.Run(context.Background(), os.Args)
}
