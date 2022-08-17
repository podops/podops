package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/podops/podops"
	"github.com/podops/podops/internal/importer"
)

// Initialize a new podcast project and repository
func ImportCommand(c *cli.Context) error {

	if c.NArg() < 1 {
		return podops.ErrInvalidNumArguments
	}

	// collect all the parameters
	feedUrl := c.Args().First()
	output := c.String("output")
	//rewrite := boolFlag(c, "rewrite")        // --rewrite
	singleFile := boolFlag(c, "single-file") // --single-file

	// get the working directory
	root := "."
	if output != "" {
		root = output
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

	show, err := importer.ImportPodcastFeed(feedUrl)
	if err != nil {
		return err
	}

	if singleFile {
		err := dumpResource(fmt.Sprintf("%s/show.yaml", root), show)
		if err != nil {
			return err
		}
	} else {
		for _, e := range show.Episodes {
			season := e.SeasonAsInt()
			episode := e.EpisodeAsInt()
			err := dumpResource(fmt.Sprintf("%s/episode-S%dE%d.yaml", root, season, episode), e)
			if err != nil {
				return err
			}
		}

		// unlink the episodes, we already dumped them
		show.Episodes = nil

		err := dumpResource(fmt.Sprintf("%s/show.yaml", root), show)
		if err != nil {
			return err
		}
	}

	return nil
}
