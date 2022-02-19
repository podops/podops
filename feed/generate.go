package feed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/txsvc/stdlib/v2/timestamp"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/loader"
)

type (
	PodcastEpisodeFrontmatter struct {
		Title   string         `json:"title" yaml:"title"`
		Slug    string         `json:"slug" yaml:"slug"`
		Date    string         `json:"date" yaml:"date"`
		Podcast PodcastEpisode `json:"podcast" yaml:"podcast"`
	}

	PodcastEpisode struct {
		MP3         string       `json:"mp3" yaml:"mp3"`
		Duration    string       `json:"duration" yaml:"duration"`
		Image       PodcastImage `json:"image" yaml:"image"`
		Episode     string       `json:"episode" yaml:"episode"`
		EpisodeType string       `json:"episodeType" yaml:"episodeType"`
		Season      string       `json:"season" yaml:"season"`
		Explicit    string       `json:"explicit" yaml:"explicit"`
		Block       string       `json:"bock" yaml:"block"`
	}

	PodcastImage struct {
		URI     string `json:"src" yaml:"src"`
		AltText string `json:"alt" yaml:"alt"`
	}
)

// Generate gathers all podcast episodes and builds the markdown assets needed to build a static site
func Generate(ctx context.Context, root, target string) error {

	// cache dir
	assetPath := filepath.Join(root, config.BuildLocation)

	// find all episodes
	now := timestamp.Now()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
					// create a *.md stub for the static website
					if err := CreateEpisodeMarkdown(assetPath, target, episode); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})

	return err
}

func CreateEpisodeMarkdown(root, target string, episode *podops.Episode) error {
	mdPath := filepath.Join(target, fmt.Sprintf("%s.md", episode.Metadata.Name))

	f, err := os.Create(mdPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// load the asset references
	enclosureAssetRef, err := LoadAssetRef(filepath.Join(root, fmt.Sprintf("%s.yaml", episode.Enclosure.AssetReference(episode.Parent()))))
	if err != nil {
		return err
	}
	imageAssetRef, err := LoadAssetRef(filepath.Join(root, fmt.Sprintf("%s.yaml", episode.Image.AssetReference(episode.Parent()))))
	if err != nil {
		return err
	}

	// create the ymal front matter
	fm := PodcastEpisodeFrontmatter{
		Title: episode.Description.Title,
		Slug:  episode.Metadata.Name,
		Date:  episode.Metadata.Date,
		Podcast: PodcastEpisode{
			MP3:      enclosureAssetRef.MediaReference(),
			Duration: ParseDuration((int64)(enclosureAssetRef.Duration)),
			Image: PodcastImage{
				URI:     imageAssetRef.MediaReference(),
				AltText: episode.Description.Title,
			},
			Episode:     episode.Metadata.Labels[podops.LabelEpisode],
			EpisodeType: episode.Metadata.Labels[podops.LabelType],
			Season:      episode.Metadata.Labels[podops.LabelSeason],
			Explicit:    episode.Metadata.Labels[podops.LabelExplicit],
			Block:       episode.Metadata.Labels[podops.LabelBlock],
		},
	}

	fmb, err := yaml.Marshal(&fm)
	if err != nil {
		return err
	}

	// write the file
	f.WriteString("---\n")
	f.Write(fmb)
	f.WriteString("---\n")

	f.WriteString(episode.Description.EpisodeText)

	return nil
}
