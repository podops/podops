package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/podops/podops"
	"github.com/podops/podops/client"
	"github.com/podops/podops/config"

	//"github.com/podops/podops/feed"
	"github.com/podops/podops/internal/builder"
	"github.com/podops/podops/internal/loader"
)

// BuildCommand builds the podcast feed
func BuildCommand(c *cli.Context) error {

	if c.NArg() > 1 {
		return podops.ErrInvalidNumArguments
	}

	root, err := ResolveRootDirectory(c)
	if err != nil {
		return err
	}

	validateOnly := boolFlag(c, "validate") // --validate
	cleanFirst := boolFlag(c, "clean")      // --clean
	buildOnly := boolFlag(c, "build-only")  // --build-only

	name, err := builder.Build(context.TODO(), root, validateOnly, buildOnly, cleanFirst)
	if err != nil {
		return err
	}

	printMsg(podops.MsgBuildSuccess, name)
	return nil
}

// AssembleCommand collect all podcast assets
func AssembleCommand(c *cli.Context) error {

	if c.NArg() > 1 {
		return podops.ErrInvalidNumArguments
	}

	root, err := ResolveRootDirectory(c)
	if err != nil {
		return err
	}

	force := boolFlag(c, "force") // --force

	err = builder.Assemble(context.TODO(), root, force)
	if err != nil {
		return err
	}

	printMsg(podops.MsgAssembleSuccess)
	return nil
}

func SyncCommand(c *cli.Context) error {

	if c.NArg() > 1 {
		return podops.ErrInvalidNumArguments
	}

	root := "."
	if c.NArg() == 1 {
		root = c.Args().First()
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		root = dir
	}
	purge := boolFlag(c, "purge") // --purge

	// reload the local credentials
	config.UpdateClientSettings(filepath.Join(root, config.DefaultConfigFileLocation))

	// find and load the show.yaml
	showPath := filepath.Join(root, config.DefaultShowName)
	_, kind, parent, err := loader.ReadResource(context.TODO(), showPath)
	if err != nil {
		return err
	}
	if kind != podops.ResourceShow {
		return podops.ErrResourceNotFound
	}
	if !podops.ValidGUID(parent) {
		return podops.ErrInvalidGUID
	}

	if err := client.Sync(parent, root, purge); err != nil {
		return err
	}

	printMsg(podops.MsgSyncSuccess)
	return nil
}
