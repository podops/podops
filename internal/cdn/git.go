package cdn

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/mmcdole/gofeed"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/feed"
)

func CloneOrPullRepo(repo, parent string) error {

	if parent == "" {
		return podops.ErrInvalidGUID
	}

	// swap the GUID for the short name
	feedPath := filepath.Join(config.StorageLocation, parent, feed.DefaultFeedName)

	// parse feed.xml and extract the name & uri
	file, err := os.Open(feedPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fp := gofeed.NewParser()
	feed, err := fp.Parse(file)
	if err != nil {
		return err
	}
	link, err := url.Parse(feed.Link)
	if err != nil {
		return err
	}
	parts := strings.Split(link.Path, "/")
	name := parts[len(parts)-1]

	repoPath := filepath.Join(config.StaticLocation, name)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// git clone
		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:               repo,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			return err
		}
	} else {
		// git pull
		r, err := git.PlainOpen(repoPath)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil {
			return err
		}
	}
	return nil
}
