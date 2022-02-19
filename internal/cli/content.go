package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/txsvc/stdlib/v2/validate"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/feed"
	"github.com/podops/podops/internal"
)

const (
	gitIgnoreText = ".build/**\n**/.podops/config\n\n// Uncomment the follwing lines if you want keep your media assets in git\n\n// *.mp3\n// *.png\n"
)

// Initialize a new podcast project and repository
func NewCommand(c *cli.Context) error {

	if c.NArg() > 1 {
		return podops.ErrInvalidNumArguments
	}

	root, err := ResolveRootDirectory(c)
	if err != nil {
		return err
	}

	// create the default .gitignore file
	ignoreFilePath := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
		f, err := os.Create(ignoreFilePath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.WriteString(gitIgnoreText)
		if err != nil {
			return err
		}
	}

	// setup the default show.yaml
	showFilePath := filepath.Join(root, feed.DefaultShowName)
	if _, err := os.Stat(showFilePath); os.IsNotExist(err) {
		guid := internal.CreateRandomAssetGUID()
		show := podops.DefaultShow("NAME", "TITLE", "SUMMARY", guid, config.Settings().GetOption(config.PodopsServiceEndpointEnv), config.Settings().GetOption(config.PodopsContentEndpointEnv))

		if err := dumpResource(showFilePath, show); err != nil {
			return err
		}
	}

	return nil
}

// TemplateCommand creates a resource template with all default values
func TemplateCommand(c *cli.Context) error {
	template := c.Args().First()

	if c.NArg() < 1 {
		return podops.ErrMissingResourceName
	}

	if !validate.IsMemberOf(template, podops.ResourceShow, podops.ResourceEpisode, podops.ResourceAsset) {
		return fmt.Errorf(podops.MsgResourceUnsupportedKind, template)
	}

	guid := c.String("guid")
	if guid == "" {
		guid = internal.CreateRandomAssetGUID()
	}
	parentGUID := c.String("parent")
	if parentGUID == "" {
		parentGUID = "PARENT-GUID"
	}
	parentName := "PARENT-NAME"

	name := strings.ToLower(fmt.Sprintf("%s-%s", template, guid))
	if c.NArg() == 2 {
		name = c.Args().Get(1)
	}

	if !podops.ValidName(name) {
		return podops.ErrInvalidResourceName
	}

	// create the yamls
	if template == podops.ResourceShow {
		show := podops.DefaultShow(name, "TITLE", "SUMMARY", guid, config.Settings().GetOption(config.PodopsServiceEndpointEnv), config.Settings().GetOption(config.PodopsContentEndpointEnv))
		err := dumpResource(fmt.Sprintf("show-%s.yaml", guid), show)
		if err != nil {
			return err
		}
	} else if template == podops.ResourceEpisode {
		episode := podops.DefaultEpisode(name, parentName, guid, parentGUID, config.Settings().GetOption(config.PodopsServiceEndpointEnv), config.Settings().GetOption(config.PodopsContentEndpointEnv))
		err := dumpResource(fmt.Sprintf("episode-%s.yaml", guid), episode)
		if err != nil {
			return err
		}
	}

	return nil
}
