package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/txsvc/stdlib/v2/id"
	"github.com/txsvc/stdlib/v2/settings"

	"github.com/podops/podops"
	"github.com/podops/podops/auth"
	"github.com/podops/podops/client"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal"
	"github.com/podops/podops/internal/loader"
)

// Initialize the CDN
func InitCommand(c *cli.Context) error {

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

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return podops.ErrInvalidParameters
	}

	// register the repo now
	return register(config.Settings().Credentials.UserID, root)
}

// Create a default config for the CDN and API services
func ConfigCommand(c *cli.Context) error {

	if c.NArg() > 2 {
		return podops.ErrInvalidNumArguments
	}

	root := ""
	phrase := ""

	if c.NArg() == 1 {
		root = c.Args().First() // path only
	} else if c.NArg() == 2 {
		root = c.Args().First() // path and pass phrase
		phrase = c.Args().Get(1)
	} else {
		root, _ = ResolveRootDirectory(c) // nothing was provided
	}

	mnemonic, err := internal.CreateMnemonic(phrase)
	if err != nil {
		return err
	}

	cfg, _ := config.LoadClientSettings(root)
	cfg.Credentials = &settings.Credentials{
		ProjectID: config.ProjectName,
		UserID:    id.Fingerprint(mnemonic),
		Token:     internal.CreateSimpleToken(),
		Expires:   0,
	}
	cfg.APIKey = cfg.Credentials.UserID
	cfg.Scopes = []string{auth.ScopeAdmin}

	configFile := filepath.Join(root, config.DefaultConfigFileLocation)
	if err := cfg.WriteToFile(configFile); err != nil {
		return err
	}

	fmt.Printf("\nuser-id: %s\n", cfg.Credentials.UserID)
	if phrase == "" {
		fmt.Printf("passphrase: %s\n\n", mnemonic)
		fmt.Println("Make a copy of the pass phrase and keep it secure !")
	}

	return nil
}

func register(userID, root string) error {
	// try to read the parent GUID from the show.yaml
	showPath := filepath.Join(root, config.DefaultShowName)

	_, kind, parent, err := loader.ReadResource(context.TODO(), showPath)
	if err != nil {
		return err
	}
	if kind != podops.ResourceShow {
		return podops.ErrBuildNoShow
	}

	// try to register the repo
	cfg, err := client.Init(userID, parent)
	if err != nil {
		return err
	}

	// write the credentials to the local repo
	credPath := filepath.Join(root, config.DefaultConfigFileLocation)
	return cfg.WriteToFile(credPath)
}
