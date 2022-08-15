package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/txsvc/stdlib/v2/timestamp"
	"github.com/txsvc/stdlib/v2/validate"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/loader"
	"github.com/podops/podops/internal/rss"
)

var (
	mediaTypeMap map[string]rss.EnclosureType
)

type (
	// EpisodeList holds the list of valid episodes that will be added to a podcast
	EpisodeList []*podops.Episode
)

func (e EpisodeList) Len() int      { return len(e) }
func (e EpisodeList) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e EpisodeList) Less(i, j int) bool {
	return e[i].PublishDateTimestamp() > e[j].PublishDateTimestamp() // sorting direction is descending
}

func init() {
	mediaTypeMap = make(map[string]rss.EnclosureType)
	mediaTypeMap["audio/x-m4a"] = rss.M4A
	mediaTypeMap["video/x-m4v"] = rss.M4V
	mediaTypeMap["video/mp4"] = rss.MP4
	mediaTypeMap["audio/mpeg"] = rss.MP3
	mediaTypeMap["video/quicktime"] = rss.MOV
	mediaTypeMap["application/pdf"] = rss.PDF
	mediaTypeMap["document/x-epub"] = rss.EPUB
}

// Build gathers all podcast resources and builds the feed.xml
func Build(ctx context.Context, root string, validateOnly, buildOnly, purge bool) (string, error) {
	var episodes EpisodeList
	episodeLookup := make(map[string]*podops.Episode)
	episodePath := make(map[string]string)

	// cache dir
	assetPath := filepath.Join(root, config.BuildLocation)

	// clean cache dir?
	if purge {
		os.RemoveAll(assetPath)
	}
	// create cache dir if needed
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		os.Mkdir(assetPath, os.ModePerm)
	}

	// find and load the show.yaml
	showPath := filepath.Join(root, config.DefaultShowName)
	rsrc, kind, parentGUID, err := loader.ReadResource(ctx, showPath)
	if err != nil {
		return "", err
	}
	if kind != podops.ResourceShow {
		return "", podops.ErrBuildNoShow
	}

	// convert and validate show.yaml
	show := rsrc.(*podops.Show)

	// validate the show
	v := show.Validate("show."+parentGUID, validate.NewValidator())
	if v.Errors != 0 {
		return show.Metadata.Name, v.AsError()
	}
	// validate show assets
	if err = ValidateResource(ctx, parentGUID, root, &show.Image); err != nil {
		return show.Metadata.Name, err
	}
	if !buildOnly {
		if err := ResolveResource(ctx, parentGUID, root, false, &show.Image); err != nil {
			return show.Metadata.Name, err
		}
	}

	// find all episodes
	now := timestamp.Now()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		base, _ := filepath.Split(path)

		// skip the assets dir
		if filepath.Base(base) == config.BuildLocation {
			return nil
		}

		if filepath.Ext(path) == ".yaml" {
			rsrc, kind, _, err := loader.ReadResource(ctx, path)
			if err != nil {
				return err
			}
			if kind == podops.ResourceEpisode {
				episode := rsrc.(*podops.Episode)
				if episode.PublishDateTimestamp() < now { // episodes with a FUTURE timestamp are valid but will be excluded
					v = episode.Validate("episode."+episode.GUID(), v)

					// FIXME filter for other flags, e.g. Block = true

					if v.Errors == 0 {
						// FIXME more validations

						// check mismatch episode.parent & show.guid
						if episode.Metadata.Parent != "" && episode.Metadata.Parent != show.Metadata.GUID {
							v.AddError(fmt.Sprintf(podops.MsgResourceInvalidReference, episode.Metadata.Parent))
						}

						// add to lookup structure
						episodePath[episode.Metadata.GUID] = base
						episodeLookup[episode.Metadata.GUID] = episode

						// add to list of episodes
						episodes = append(episodes, episode)
					}
				}
			}
		}
		return nil
	})

	// abort here in case of any errors so far ...

	if err != nil {
		return show.Metadata.Name, err
	}
	if v.Errors != 0 {
		return show.Metadata.Name, v.AsError()
	}
	if len(episodes) == 0 {
		return show.Metadata.Name, podops.ErrBuildNoEpisodes
	}

	// sort episodes, descending by timestamp
	sort.Sort(episodes)

	// assemble the feed
	feed, err := transformToPodcast(show)
	if err != nil {
		return show.Metadata.Name, err
	}
	tt, _ := time.Parse(time.RFC1123Z, episodes[0].PublishDate())
	feed.AddPubDate(&tt)

	// add all published episodes
	for _, e := range episodes {
		if err = ValidateResource(ctx, parentGUID, root, &e.Image); err != nil {
			return show.Metadata.Name, err
		}
		if !buildOnly {
			if err := ResolveResource(ctx, parentGUID, root, false, &e.Image); err != nil {
				return show.Metadata.Name, err
			}
		}

		if err = ValidateResource(ctx, parentGUID, root, &e.Enclosure); err != nil {
			return show.Metadata.Name, err
		}
		if !buildOnly {
			if err := ResolveResource(ctx, parentGUID, root, false, &e.Enclosure); err != nil {
				return show.Metadata.Name, err
			}
		}

		item, err := transformToItem(e)
		if err != nil {
			return show.Metadata.Name, err
		}
		feed.AddItem(item)
	}

	if validateOnly {
		return show.Metadata.Name, nil
	}

	// write the feed.xml
	feedPath := filepath.Join(root, config.BuildLocation, config.DefaultFeedName)
	return show.Metadata.Name, os.WriteFile(feedPath, feed.Bytes(), 0644)

}

// Assemble collects all referenced podcast resources (.mp3, .gif, .png etc)
// and puts them into the local build location.
func Assemble(ctx context.Context, root string, force bool) error {

	// FIXME flag purge does nothing

	// find and load the show.yaml
	showPath := filepath.Join(root, config.DefaultShowName)
	_, kind, parent, err := loader.ReadResource(ctx, showPath)
	if err != nil {
		return err
	}
	if kind != podops.ResourceShow {
		return podops.ErrBuildNoShow
	}

	// cache dir
	assetPath := filepath.Join(root, config.BuildLocation)
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		return podops.ErrAssembleNoResources
	}

	err = filepath.Walk(assetPath, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)

		if ext == ".yaml" {
			encl, err := loader.ReadEnclosure(ctx, path)
			if err != nil {
				return err
			}
			return ResolveResource(ctx, parent, root, force, encl)
		}

		// ignore all other extensions
		return nil
	})

	return err
}
