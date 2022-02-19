package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
)

var (
	MsgMissingCmdParameters = "missing parameter(s). try 'po help %s'"
	MsgInvalidCmdParameters = "unknown sub-command '%s'. try 'po help %s'"
)

// NoOpCommand is just a placeholder
func NoOpCommand(c *cli.Context) error {
	return cli.Exit(fmt.Sprintf("Command '%s' is not implemented", c.Command.Name), 0)
}

func InfoCommand(c *cli.Context) error {
	data, err := yaml.Marshal(config.Settings())
	if err != nil {
		return err
	}

	fmt.Printf("path: %s\n---\n%s\n", config.SettingsPath(), string(data))
	return nil
}

func ResolveRootDirectory(c *cli.Context) (string, error) {
	root := "."
	if c.NArg() == 1 {
		root = c.Args().First()
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return root, err
		}
		root = dir
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return root, podops.ErrInvalidParameters
	}

	return root, nil
}

func boolFlag(c *cli.Context, flag string) bool {
	value := c.String(flag)

	if value != "" {
		if strings.ToLower(value) == "true" || strings.ToLower(value) == "yes" {
			return true
		}
	}
	return false
}

func dumpResource(path string, doc interface{}) error {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	ioutil.WriteFile(path, data, 0644)
	fmt.Printf("\n---\n# %s\n%s\n", path, string(data))

	return nil
}

// printMsg is used for all the cli output
func printMsg(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}
