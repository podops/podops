package importer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal"
)

const (
	parseRSSTimeout = 60 * time.Second
)

func ImportPodcastFeed(feedUrl string) (*podops.Show, error) {
	ctx, cancel := context.WithTimeout(context.Background(), parseRSSTimeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.UserAgent = config.UserAgentString

	feed, err := fp.ParseURLWithContext(feedUrl, ctx)
	if err != nil {
		return nil, err
	}

	// import the description of the show
	show := importShow(feed)

	// import all episodes
	fixEpisodeNumbers := false
	countFixEpisodeNumbers := 0
	show.Episodes = make(podops.EpisodeList, len(feed.Items))
	for i, item := range feed.Items {
		e := importEpisode(item)

		if e.EpisodeAsInt() < 1 {
			// incomplete numbering ?
			countFixEpisodeNumbers++
		}
		if e.Image.URI == "" {
			// re-use the shows image for the episode
			e.Image = podops.AssetRef{
				URI: show.Image.URI,
				Rel: show.Image.Rel,
			}
		}
		if e.Metadata.Parent == "" {
			e.Metadata.Parent = show.Metadata.GUID
		}
		// finally add the episode to the array
		show.Episodes[i] = e
	}

	// sort episodes, descending by timestamp
	sort.Sort(show.Episodes)

	// only fix the numbering if ALL the episode numbers are empty !
	fixEpisodeNumbers = (countFixEpisodeNumbers == len(show.Episodes))

	if fixEpisodeNumbers {
		maxEpisode := len(show.Episodes)
		for i := range show.Episodes {
			show.Episodes[i].Metadata.Labels[podops.LabelEpisode] = fmt.Sprintf("%d", maxEpisode)
			maxEpisode--
		}
	}

	// adjust the publish date if the timestamp on the show is older than the one from the latest episode
	if show.PublishDateTimestamp() < show.Episodes[0].PublishDateTimestamp() {
		show.Metadata.Date = show.Episodes[0].Metadata.Date
	}

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
		Name:   formatName(feed.Title),
		GUID:   internal.CreateShortGUID(feed.Title),
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
		EpisodeText: stringWithDefault(item.Content, item.Description),
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
		Name:   formatName(item.Title),
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
		m.Labels[podops.LabelSeason] = stringWithDefault(item.ITunesExt.Season, "1")
		m.Labels[podops.LabelEpisode] = stringWithDefault(item.ITunesExt.Episode, "")
		m.Labels[podops.LabelExplicit] = stringExpect(item.ITunesExt.Explicit, "True", "False")
		m.Labels[podops.LabelBlock] = stringExpect(item.ITunesExt.Block, "Yes", "No")
	}

	return m
}
