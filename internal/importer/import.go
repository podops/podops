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
	srcPath := fmt.Sprintf("../../hack/%s_src.yaml", id.Fingerprint(feed.Title))
	targetPath := fmt.Sprintf("../../hack/%s.yaml", id.Fingerprint(feed.Title))
	dumpResource(srcPath, feed)
	// end hack

	show := createShow(feed)

	// hack
	dumpResource(targetPath, show)
	// end hack

	return &show, nil
}

func createShow(feed *gofeed.Feed) podops.Show {
	show := podops.Show{
		APIVersion: config.Version,
		Kind:       podops.ResourceShow,
		Metadata: podops.Metadata{
			Name:   feed.Title,
			GUID:   id.Fingerprint(feed.Title),
			Date:   feed.Updated, //time.Now().UTC().Format(time.RFC1123Z),
			Labels: podops.DefaultShowMetadata(),
		},
		Description: createShowDescription(feed),
		Image: podops.AssetRef{
			URI: feed.Image.URL,
			Rel: podops.ResourceTypeExternal,
		},
	}

	return show
}

func createShowDescription(feed *gofeed.Feed) podops.ShowDescription {
	desc := podops.ShowDescription{
		Title: feed.Title,
		// ADD SubTitle to podops.ShowDescription
		Summary: feedSummary(feed),
		Link: podops.AssetRef{
			URI: feed.Link,
			Rel: podops.ResourceTypeExternal,
		},
	}

	return desc
}

func feedSummary(feed *gofeed.Feed) string {
	if feed.ITunesExt != nil {
		return feed.ITunesExt.Summary
	}
	return ""
}
