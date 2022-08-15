package importer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/txsvc/stdlib/v2/id"
	"gopkg.in/yaml.v3"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal"
)

const (
	parseRssTimeout = 60 * time.Second
)

func dumpResource(path string, doc interface{}) error {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	os.WriteFile(path, data, 0644)
	fmt.Printf("\n---\n# %s\n%s\n", path, string(data))

	return nil
}

func ImportPodcastFeed(feedUrl string) (*podops.Show, error) {
	ctx, cancel := context.WithTimeout(context.Background(), parseRssTimeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.UserAgent = config.UserAgentString

	feed, err := fp.ParseURLWithContext(feedUrl, ctx)
	if err != nil {
		return nil, err
	}

	// hack
	srcPath := fmt.Sprintf("../../hack/import/%s_src.yaml", id.Fingerprint(feed.Title))
	targetPath := fmt.Sprintf("../../hack/import/%s.yaml", id.Fingerprint(feed.Title))
	dumpResource(srcPath, feed)
	// end hack

	// import the description of the show
	show := importShow(feed)
	//show.Episodes = importEpisodes(feed)

	// import all episodes
	show.Episodes = make(podops.EpisodeList, len(feed.Items))
	for i, item := range feed.Items {
		show.Episodes[i] = importEpisode(item)
	}

	// hack
	dumpResource(targetPath, show)
	// end hack

	return &show, nil
}

func importShow(feed *gofeed.Feed) podops.Show {
	show := podops.Show{
		APIVersion: config.Version,
		Kind:       podops.ResourceShow,
		Metadata: podops.Metadata{
			Name:   feed.Title,
			GUID:   id.Fingerprint(feed.Title),
			Labels: podops.DefaultShowMetadata(),
		},
		//Description: createShowDescription(feed),
		Image: podops.AssetRef{
			URI: feed.Image.URL,
			Rel: podops.ResourceTypeExternal,
		},
	}

	return show
}

func importEpisode(item *gofeed.Item) *podops.Episode {
	episode := podops.Episode{
		APIVersion: config.Version,
		Kind:       podops.ResourceEpisode,
		Metadata: podops.Metadata{
			Name:   item.Title,
			GUID:   itemGUID(item),
			Labels: importEpisodeLabels(item),
		},
		Description: importEpisodeDescription(item),
		Image:       importItemImageAssetRef(item),
		Enclosure:   importEnclosureAssetRef(item),
	}

	return &episode
}

func importEpisodeDescription(item *gofeed.Item) podops.EpisodeDescription {
	ed := podops.EpisodeDescription{
		Title:       item.Title,
		Summary:     item.Description,
		EpisodeText: item.Content,
		Link: podops.AssetRef{
			URI: item.Link,
			Rel: podops.ResourceTypeExternal,
		},
		Duration: 0,
	}

	// patch with maybe better data
	if item.ITunesExt != nil {
		ed.Summary = stringWithDefault(item.ITunesExt.Summary, ed.Summary)
		ed.Duration = internal.ConvTimeStringToSeconds(item.ITunesExt.Duration)
	}
	return ed
}

// importEnclosureAssetRef returns the first enclosure only !
func importEnclosureAssetRef(item *gofeed.Item) podops.AssetRef {
	ar := podops.AssetRef{
		URI:      item.Enclosures[0].URL,
		Rel:      podops.ResourceTypeExternal,
		Type:     item.Enclosures[0].Type,
		Duration: convDurationToInt(item.ITunesExt),
		Size:     internal.ConvStrToInt(item.Enclosures[0].Length),
	}

	return ar
}

// importItemImageAssetRef returns the img url, iTunesExt has priority
func importItemImageAssetRef(item *gofeed.Item) podops.AssetRef {
	var url = ""

	if item.Image != nil {
		url = item.Image.URL
	}
	if item.ITunesExt != nil {
		url = stringWithDefault(item.ITunesExt.Image, url)
	}

	return podops.AssetRef{
		URI: url,
		Rel: podops.ResourceTypeExternal,
	}
}

func importEpisodeLabels(item *gofeed.Item) map[string]string {
	l := podops.DefaultEpisodeMetadata()

	if item.ITunesExt != nil {
		l[podops.LabelSeason] = stringWithDefault(item.ITunesExt.Season, l[podops.LabelSeason])
		l[podops.LabelEpisode] = stringWithDefault(item.ITunesExt.Episode, l[podops.LabelEpisode])

		/*
			l[LabelExplicit] = "no"
			l[LabelType] = EpisodeTypeFull
			l[LabelBlock] = "no"
		*/
	}

	return l
}
