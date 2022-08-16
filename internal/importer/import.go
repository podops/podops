package importer

import (
	"context"
	"fmt"
	"os"
	"strings"
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

func dumpResource(path string, doc interface{}, dump bool) error {
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	os.WriteFile(path, data, 0644)

	if dump {
		fmt.Printf("\n---\n# %s\n%s\n", path, string(data))
	}

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
	dumpResource(srcPath, feed, false)
	// end hack

	// import the description of the show
	show := importShow(feed)

	/*
		// import all episodes
		show.Episodes = make(podops.EpisodeList, len(feed.Items))
		for i, item := range feed.Items {
			show.Episodes[i] = importEpisode(item)
		}
	*/

	// FIXME: sort episodes, assign numbers if not present, patch show.metadata.date if inconsistent

	// hack
	dumpResource(targetPath, show, true)
	// end hack

	return &show, nil
}

func importShow(feed *gofeed.Feed) podops.Show {
	show := podops.Show{
		APIVersion:  config.Version,
		Kind:        podops.ResourceShow,
		Metadata:    importShowMetadata(feed),
		Description: importShowDescription(feed),
		Image:       importShowImageAssetRef(feed),
	}

	if feed.FeedLink != "" {
		show.FeedLink = &podops.AssetRef{
			URI: feed.FeedLink,
			Rel: podops.ResourceTypeExternal,
		}
	}

	if feed.ITunesExt != nil {
		if feed.ITunesExt.NewFeedURL != "" {
			show.NewFeedLink = &podops.AssetRef{
				URI: feed.ITunesExt.NewFeedURL,
				Rel: podops.ResourceTypeExternal,
			}
		}
	}

	return show
}

func importShowMetadata(feed *gofeed.Feed) podops.Metadata {
	m := podops.Metadata{
		Name:   feed.Title,
		GUID:   id.Fingerprint(feed.Title),
		Labels: podops.DefaultShowMetadata(),
	}

	m.Labels[podops.LabelLanguage] = stringWithDefault(feed.Language, "en")

	if feed.Published != "" {
		m.Date = feed.Published
	} else if feed.Updated != "" {
		m.Date = feed.Updated
	}

	if feed.ITunesExt != nil {
		if feed.ITunesExt.Keywords != "" {
			m.Tags = feed.ITunesExt.Keywords
		}

		m.Labels[podops.LabelType] = stringWithDefault(strings.Title(feed.ITunesExt.Type), podops.ShowTypeEpisodic)
		m.Labels[podops.LabelExplicit] = stringExpect(feed.ITunesExt.Explicit, "True", "False")
		m.Labels[podops.LabelBlock] = stringExpect(feed.ITunesExt.Block, "Yes", "No")
		m.Labels[podops.LabelComplete] = stringExpect(feed.ITunesExt.Complete, "Yes", "No")
	}

	return m
}

func importShowDescription(feed *gofeed.Feed) podops.ShowDescription {
	sd := podops.ShowDescription{
		Title:   feed.Title,
		Summary: feed.Description,
		Link: podops.AssetRef{
			URI: feed.Link,
			Rel: podops.ResourceTypeExternal,
		},
		Category:  importCategory(feed),
		Owner:     importOwner(feed),
		Author:    importAuthor(feed),
		Copyright: feed.Copyright,
	}

	return sd
}

func importCategory(feed *gofeed.Feed) []podops.Category {
	var cc *podops.Category                 // current category
	cm := make(map[string]*podops.Category) // category map
	ca := make([]*podops.Category, 0)       // category array

	for _, c := range feed.Categories {
		lc, found := cm[c]
		if found {
			cc = lc
		} else {
			if strings.HasPrefix(c, " ") {
				// found a subcategory
				if len(cc.SubCategory) == 0 {
					cc.SubCategory = make([]string, 1)
					cc.SubCategory[0] = strings.Trim(c, " ")
				} else {
					cc.SubCategory = append(cc.SubCategory, strings.Trim(c, " "))
				}
			} else {
				// new category, add to map and array
				cc = &podops.Category{
					Name: c,
				}
				ca = append(ca, cc) // append to array
				cm[c] = cc
			}
		}
	}

	// copy from *structs to structs
	categories := make([]podops.Category, len(ca))
	for i, cc := range ca {
		categories[i] = podops.Category{
			Name:        cc.Name,
			SubCategory: cc.SubCategory,
		}
	}
	return categories
}

func importOwner(feed *gofeed.Feed) podops.Owner {
	if feed.ITunesExt != nil && feed.ITunesExt.Owner != nil {
		return podops.Owner{
			Name:  feed.ITunesExt.Owner.Name,
			Email: feed.ITunesExt.Owner.Email,
		}
	}
	return podops.Owner{
		Name:  feed.Author.Name,
		Email: feed.Author.Email,
	}
}

func importAuthor(feed *gofeed.Feed) string {
	if feed.ITunesExt != nil && feed.ITunesExt.Author != "" {
		return feed.ITunesExt.Author
	}
	return feed.Author.Name
}

// importShowImageAssetRef returns the img url, iTunesExt has priority
func importShowImageAssetRef(feed *gofeed.Feed) podops.AssetRef {
	var url = ""

	if feed.Image != nil {
		url = feed.Image.URL
	}
	if feed.ITunesExt != nil {
		url = stringWithDefault(feed.ITunesExt.Image, url)
	}

	return podops.AssetRef{
		URI: url,
		Rel: podops.ResourceTypeExternal,
	}
}

func importEpisode(item *gofeed.Item) *podops.Episode {
	episode := podops.Episode{
		APIVersion:  config.Version,
		Kind:        podops.ResourceEpisode,
		Metadata:    importEpisodeMetadata(item),
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

func importEpisodeMetadata(item *gofeed.Item) podops.Metadata {
	m := podops.Metadata{
		Name:   item.Title,
		GUID:   itemGUID(item),
		Labels: podops.DefaultEpisodeMetadata(),
	}

	if item.Published != "" {
		m.Date = item.Published
	} else if item.Updated != "" {
		m.Date = item.Updated
	}

	if item.ITunesExt != nil {
		if item.ITunesExt.Keywords != "" {
			m.Tags = item.ITunesExt.Keywords
		}

		m.Labels[podops.LabelType] = stringWithDefault(strings.Title(item.ITunesExt.EpisodeType), podops.EpisodeTypeFull)
		m.Labels[podops.LabelSeason] = stringWithDefault(item.ITunesExt.Season, m.Labels[podops.LabelSeason])
		m.Labels[podops.LabelEpisode] = stringWithDefault(item.ITunesExt.Episode, m.Labels[podops.LabelEpisode])
		m.Labels[podops.LabelExplicit] = stringExpect(item.ITunesExt.Explicit, "True", "False")
		m.Labels[podops.LabelBlock] = stringExpect(item.ITunesExt.Block, "Yes", "No")
	}

	return m
}
