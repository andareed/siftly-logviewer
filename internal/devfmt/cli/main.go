package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/devfmt"
	"github.com/urfave/cli/v3"
)

func Run() {
	app := &cli.Command{
		Name:    "devfmt",
		Usage:   "Siftly Toolkit: Device Info Viewer",
		Version: devfmt.Version,
		Action: func(_ context.Context, c *cli.Command) error {
			if c.Bool("version") {
				fmt.Println("Version:", devfmt.Version)
				return nil
			}
			return runExportLike(c)
		},
		Flags: commonFlags(),
		Commands: []*cli.Command{
			{
				Name:  "list-groups",
				Usage: "Output distinct group_category values seen in input",
				Flags: commandInputFlags(),
				Action: func(_ context.Context, c *cli.Command) error {
					inputPath := resolveInputPath(c)
					ds, _, err := devfmt.LoadDataset(inputPath)
					if err != nil {
						return err
					}
					for _, g := range devfmt.GroupsSeen(ds) {
						fmt.Println(g)
					}
					return nil
				},
			},
			{
				Name:  "list-categories",
				Usage: "Output distinct raw category values seen in input",
				Flags: commandInputFlags(),
				Action: func(_ context.Context, c *cli.Command) error {
					inputPath := resolveInputPath(c)
					ds, _, err := devfmt.LoadDataset(inputPath)
					if err != nil {
						return err
					}
					for _, cat := range devfmt.CategoriesSeen(ds) {
						fmt.Println(cat)
					}
					return nil
				},
			},
			{
				Name:  "show",
				Usage: "Lookup and render one entity by raw category and id",
				Flags: append(commandInputFlags(),
					&cli.StringFlag{Name: "category", Required: true, Usage: "Raw category exact match"},
					&cli.StringFlag{Name: "id", Required: true, Usage: "Entity id exact match"},
					&cli.StringFlag{Name: "search", Usage: "Case-insensitive contains match on field_value_lc"},
					&cli.BoolFlag{Name: "sort", Usage: "Sort entities by id within group"},
				),
				Action: func(_ context.Context, c *cli.Command) error {
					inputPath := resolveInputPath(c)
					q := devfmt.Query{
						Category: c.String("category"),
						ID:       c.String("id"),
						Search:   c.String("search"),
						SortID:   c.Bool("sort"),
					}
					return devfmt.Run(inputPath, c.String("debug"), q)
				},
			},
			{
				Name:  "export",
				Usage: "Render selected group in siftly",
				Flags: append(commandInputFlags(),
					&cli.StringFlag{Name: "group", Usage: "group_category exact match"},
					&cli.StringFlag{Name: "category", Usage: "Raw category exact match (debug)"},
					&cli.StringFlag{Name: "id", Usage: "Entity id exact match"},
					&cli.StringFlag{Name: "search", Usage: "Case-insensitive contains match on field_value_lc"},
					&cli.BoolFlag{Name: "sort", Usage: "Sort entities by id within group"},
				),
				Action: func(_ context.Context, c *cli.Command) error {
					return runExportLike(c)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runExportLike(c *cli.Command) error {
	inputPath := resolveInputPath(c)
	if inputPath == "" && isTTY(os.Stdin) {
		return cli.Exit("provide --input <dump> when stdin is a TTY", 1)
	}
	q := devfmt.Query{
		Group:    c.String("group"),
		Category: c.String("category"),
		ID:       c.String("id"),
		Search:   c.String("search"),
		SortID:   c.Bool("sort"),
	}

	if q.Group == "" {
		if !isTTY(os.Stdin) {
			return cli.Exit("--group is required when stdin is not a TTY", 1)
		}

		ds, _, err := devfmt.LoadDataset(inputPath)
		if err != nil {
			return err
		}
		group, err := promptGroup(devfmt.GroupsSeen(ds))
		if err != nil {
			return err
		}
		q.Group = group
	}

	return devfmt.Run(inputPath, c.String("debug"), q)
}

func isTTY(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func promptGroup(groups []string) (string, error) {
	if len(groups) == 0 {
		return "", fmt.Errorf("no groups found in input")
	}
	fmt.Println("Select group_category:")
	for i, g := range groups {
		fmt.Printf("%d) %s\n", i+1, g)
	}
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("no selection made")
	}
	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > len(groups) {
		return "", fmt.Errorf("invalid selection %q", line)
	}
	return groups[idx-1], nil
}

func commonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "debug",
			Usage: "Write debug logs to file",
		},
		&cli.StringFlag{
			Name:    "input",
			Aliases: []string{"i"},
			Usage:   "Path to input dump file (use - or omit to read stdin)",
		},
		&cli.StringFlag{Name: "group", Usage: "group_category exact match"},
		&cli.StringFlag{Name: "category", Usage: "Raw category exact match (debug)"},
		&cli.StringFlag{Name: "id", Usage: "Entity id exact match"},
		&cli.StringFlag{Name: "search", Usage: "Case-insensitive contains match on field_value_lc"},
		&cli.BoolFlag{Name: "sort", Usage: "Sort entities by id within group"},
		&cli.BoolFlag{Name: "version", Usage: "Print version and exit"},
	}
}

func commandInputFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "debug",
			Usage: "Write debug logs to file",
		},
		&cli.StringFlag{
			Name:    "input",
			Aliases: []string{"i"},
			Usage:   "Path to input dump file (use - or omit to read stdin)",
		},
	}
}

func resolveInputPath(c *cli.Command) string {
	inputPath := c.String("input")
	if inputPath == "" {
		inputPath = c.Args().First()
	}
	return inputPath
}
