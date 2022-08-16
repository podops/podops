package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/podops/podops/config"
	cmd "github.com/podops/podops/internal/cli"
)

const (
	adminCommandsGroup   = "\nAdmin Commands"
	basicCommandsGroup   = "\nBasic Commands"
	contentCommandsGroup = "\nContent Commands"
)

func main() {

	// initialize CLI
	app := &cli.App{
		Name:      config.CmdLineName,
		Version:   config.VersionString,
		Usage:     fmt.Sprintf("PodOps: Podcast Operations CLI (%s)", config.Version),
		Copyright: config.CopyrightString,
		Commands:  setupCommands(),
		Flags:     globalFlags(),
		Before: func(c *cli.Context) error {
			// handle global config
			if path := c.String("config"); path != "" {
				config.UpdateClientSettings(path)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			fmt.Println(globalHelpText)
			return nil
		},
		OnUsageError: func(context *cli.Context, err error, isSubcommand bool) error {
			fmt.Println("OnUsageError")
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %v\n", strings.ToLower(err.Error()))
		fmt.Printf("Run '%s help' for usage.\n", config.CmdLineName)
		os.Exit(1)
	}
}

func setupCommands() []*cli.Command {
	c := []*cli.Command{
		// basic commands
		{
			Name:      "build",
			Usage:     "Build the podcast feed",
			UsageText: "build [path]",
			Category:  basicCommandsGroup,
			Action:    cmd.BuildCommand,
			Flags:     buildFlags(),
		},
		{
			Name:      "assemble",
			Usage:     "Collect all podcast resources",
			UsageText: "assemble [path]",
			Category:  basicCommandsGroup,
			Action:    cmd.AssembleCommand,
			Flags:     assembleFlags(),
		},
		{
			Name:      "sync",
			Usage:     "Sync podcast repository to the CDN",
			UsageText: "sync [path]",
			Category:  basicCommandsGroup,
			Action:    cmd.SyncCommand,
			Flags:     syncFlags(),
		},
		{
			Name:      "init",
			Usage:     "Initialize the CDN",
			UsageText: "init [path]",
			Category:  basicCommandsGroup,
			Action:    cmd.InitCommand,
		},

		// content commands
		{
			Name:      "new",
			Usage:     "Initialize a new podcast project and repository",
			UsageText: "new [path]",
			Category:  contentCommandsGroup,
			Action:    cmd.NewCommand,
		},
		{
			Name:      "import",
			Usage:     "Import a podcast from its RSS feed",
			UsageText: "import [feed]",
			Category:  contentCommandsGroup,
			Action:    cmd.ImportCommand,
			Flags:     importFlags(),
		},
		{
			Name:      "template",
			Usage:     "Create resource templates with default example values",
			UsageText: "template show|episode [NAME]",
			Category:  contentCommandsGroup,
			Action:    cmd.TemplateCommand,
			Flags:     templateFlags(),
		},

		// admin commands
		{
			Name:      "info",
			Usage:     "Display system information",
			UsageText: "info",
			Category:  adminCommandsGroup,
			Action:    cmd.InfoCommand,
		},
		{
			Name:      "config",
			Usage:     "Create a default config for the CDN and API services",
			UsageText: "config [path [pass phrase]]",
			Category:  adminCommandsGroup,
			Action:    cmd.ConfigCommand,
		},
	}
	return c
}

func globalFlags() []cli.Flag {
	f := []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Usage:       "Directory for the configuration and keystore",
			DefaultText: "$HOME/.podops",
			Aliases:     []string{"c"},
		},
	}
	return f
}

func importFlags() []cli.Flag {
	f := []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Usage:       "The output location",
			DefaultText: ".",
			Aliases:     []string{"o"},
		},
		&cli.BoolFlag{
			Name:  "rewrite",
			Usage: "Rewrite asset references during the import",
		},
		&cli.BoolFlag{
			Name:  "single-file",
			Usage: "Write all resources as one file",
		},
	}
	return f
}

func templateFlags() []cli.Flag {
	f := []cli.Flag{
		&cli.StringFlag{
			Name:    "guid",
			Usage:   "The asset's unique reference ID (GUID)",
			Aliases: []string{"g"},
		},
		&cli.StringFlag{
			Name:    "parent",
			Usage:   "The asset's unique parent reference ID (Parent GUID)",
			Aliases: []string{"p"},
		},
	}
	return f
}

func buildFlags() []cli.Flag {
	f := []cli.Flag{
		&cli.BoolFlag{
			Name:    "validate",
			Usage:   "Validate the build only",
			Aliases: []string{"v"},
		},
		&cli.BoolFlag{
			Name:    "purge",
			Usage:   "Delete all files in the .build folder",
			Aliases: []string{"p"},
		},
		&cli.BoolFlag{
			Name:    "build-only",
			Usage:   "Only build the feed without collecting the podcast resources",
			Aliases: []string{"b"},
		},
	}
	return f
}

func assembleFlags() []cli.Flag {
	f := []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Usage:   "Force download of podcast resources",
			Aliases: []string{"f"},
		},
	}
	return f
}

func syncFlags() []cli.Flag {
	f := []cli.Flag{
		&cli.BoolFlag{
			Name:    "purge",
			Usage:   "Purge unused resources from the CDN",
			Aliases: []string{"p"},
		},
	}
	return f
}

// all the help texts used in the CLI
const (
	globalHelpText = `PodOps: Podcast Operations Client

This client tool helps you to create and produce podcasts.
It also includes administrative commands for managing your live podcasts.

To see the full list of supported commands, run 'po help'`
)
