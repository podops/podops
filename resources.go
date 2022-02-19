package podops

import (
	"fmt"
	"strings"
	"time"

	"github.com/podops/podops/config"
)

const (
	//
	// Required and optional labels:
	//
	//	show:
	//		language:	<ISO639 two-letter-code> REQUIRED 'channel.language'
	//		explicit:	True | False REQUIRED 'channel.itunes.explicit'
	//		type:		Episodic | Serial REQUIRED 'channel. itunes.type'
	//		block:		Yes OPTIONAL 'channel.itunes.block' Anything else than 'Yes' has no effect
	//		complete:	Yes OPTIONAL 'channel.itunes.complete' Anything else than 'Yes' has no effect
	//
	//	episode:
	//		guid:		<unique id> 'item.guid'
	//		date:		<publish date> REQUIRED 'item.pubDate'
	//		season: 	<season number> OPTIONAL 'item.itunes.season'
	//		episode:	<episode number> REQUIRED 'item.itunes.episode'
	//		explicit:	True | False REQUIRED 'channel.itunes.explicit'
	//		type:		Full | Trailer | Bonus REQUIRED 'item.itunes.episodeType'
	//		block:		Yes OPTIONAL 'item.itunes.block' Anything else than 'Yes' has no effect
	//

	// LabelLanguage ISO-639 two-letter language code. channel.language
	LabelLanguage = "language"
	// LabelExplicit ["true"|"false"] channel.itunes.explicit
	LabelExplicit = "explicit"
	// LabelType ["Episodic"|"Serial"] channel.itunes.type
	LabelType = "type"
	// LabelBlock ["Yes"] channel.itunes.block
	LabelBlock = "block"
	// LabelComplete ["Yes"] channel.itunes.complete
	LabelComplete = "complete"
	// LabelSeason defaults to "1"
	LabelSeason = "season"
	// LabelEpisode positive integer 1..
	LabelEpisode = ResourceEpisode

	// ShowTypeEpisodic type of podcast is episodic
	ShowTypeEpisodic = "Episodic"
	// ShowTypeSerial type of podcast is serial
	ShowTypeSerial = "Serial"

	// EpisodeTypeFull type of episode is 'full'
	EpisodeTypeFull = "Full"
	// EpisodeTypeTrailer type of episode is 'trailer'
	EpisodeTypeTrailer = "Trailer"
	// EpisodeTypeBonus type of episode is 'bonus'
	EpisodeTypeBonus = "Bonus"

	// ResourceTypeExternal references an external URL
	ResourceTypeExternal = "external"
	// ResourceTypeLocal references a local resource
	ResourceTypeLocal = "local"
	// ResourceTypeImport references an external resources that will be imported into the CDN
	ResourceTypeImport = "import"

	// ResourceShow is referencing a resource of type "show"
	ResourceShow = "show"
	// ResourceEpisode is referencing a resource of type "episode"
	ResourceEpisode = "episode"
	// ResourceAsset is referencing any media or binary resource e.g. .mp3 or .png
	ResourceAsset = "asset"
	// ResourceALL is a wildcard for any kind of resource
	ResourceALL = "all"
)

type (
	// Apple Podcast: https://help.apple.com/itc/podcasts_connect/#/itcb54353390
	// RSS 2.0: https://cyber.harvard.edu/rss/rss.html

	// Metadata contains information describing a resource
	Metadata struct {
		Name   string            `json:"name" yaml:"name" binding:"required"`       // REQUIRED
		GUID   string            `json:"guid" yaml:"guid" binding:"required"`       // REQUIRED
		Parent string            `json:"parent,omitempty" yaml:"parent,omitempty" ` // OPTIONAL
		Date   string            `json:"date,omitempty" yaml:"date,omitempty" `     // RECOMMENDED
		Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`  // REQUIRED
		Tags   string            `json:"tags,omitempty" yaml:"tags,omitempty" `     // OPTIONAL
	}

	// GenericResource holds only the kind and metadata of a resource
	GenericResource struct {
		APIVersion string   `json:"apiVersion" yaml:"apiVersion" binding:"required"` // REQUIRED default: v1.0
		Kind       string   `json:"kind" yaml:"kind" binding:"required"`             // REQUIRED default: show
		Metadata   Metadata `json:"metadata" yaml:"metadata" binding:"required"`     // REQUIRED
	}

	// Show holds all metadata related to a podcast/show
	Show struct {
		APIVersion  string          `json:"apiVersion" yaml:"apiVersion" binding:"required"`   // REQUIRED default: v1.0
		Kind        string          `json:"kind" yaml:"kind" binding:"required"`               // REQUIRED default: show
		Metadata    Metadata        `json:"metadata" yaml:"metadata" binding:"required"`       // REQUIRED
		Description ShowDescription `json:"description" yaml:"description" binding:"required"` // REQUIRED
		Image       AssetRef        `json:"image" yaml:"image" binding:"required"`             // REQUIRED 'channel.itunes.image'
	}

	// Episode holds all metadata related to a podcast episode
	Episode struct {
		APIVersion  string             `json:"apiVersion" yaml:"apiVersion" binding:"required"`   // REQUIRED default: v1.0
		Kind        string             `json:"kind" yaml:"kind" binding:"required"`               // REQUIRED default: episode
		Metadata    Metadata           `json:"metadata" yaml:"metadata" binding:"required"`       // REQUIRED
		Description EpisodeDescription `json:"description" yaml:"description" binding:"required"` // REQUIRED
		Image       AssetRef           `json:"image" yaml:"image" binding:"required"`             // REQUIRED 'item.itunes.image'
		Enclosure   AssetRef           `json:"enclosure" yaml:"enclosure" binding:"required"`     // REQUIRED
	}

	// ShowDescription holds essential show metadata
	ShowDescription struct {
		Title     string     `json:"title" yaml:"title" binding:"required"`          // REQUIRED 'channel.title' 'channel.itunes.title'
		Summary   string     `json:"summary" yaml:"summary" binding:"required"`      // REQUIRED 'channel.description'
		Link      AssetRef   `json:"link" yaml:"link"`                               // RECOMMENDED 'channel.link'
		Category  []Category `json:"category" yaml:"category" binding:"required"`    // REQUIRED channel.category
		Owner     Owner      `json:"owner" yaml:"owner"`                             // RECOMMENDED 'channel.itunes.owner'
		Author    string     `json:"author" yaml:"author"`                           // RECOMMENDED 'channel.itunes.author'
		Copyright string     `json:"copyright,omitempty" yaml:"copyright,omitempty"` // OPTIONAL 'channel.copyright'
		NewFeed   *AssetRef  `json:"newFeed,omitempty" yaml:"newFeed,omitempty"`     // OPTIONAL channel.itunes.new-feed-url -> move to label
	}

	// EpisodeDescription holds essential episode metadata
	EpisodeDescription struct {
		Title       string   `json:"title" yaml:"title" binding:"required"`                                 // REQUIRED 'item.title' 'item.itunes.title'
		Summary     string   `json:"summary" yaml:"summary" binding:"required"`                             // REQUIRED 'item.description'
		EpisodeText string   `json:"episodeText,omitempty" yaml:"episodeText,omitempty" binding:"required"` // REQUIRED 'item.itunes.summary'
		Link        AssetRef `json:"link" yaml:"link"`                                                      // RECOMMENDED 'item.link'
		Duration    int      `json:"duration" yaml:"duration" binding:"required"`                           // REQUIRED 'item.itunes.duration'
	}

	// Owner describes the owner of the show/podcast
	Owner struct {
		Name  string `json:"name" yaml:"name" binding:"required"`   // REQUIRED
		Email string `json:"email" yaml:"email" binding:"required"` // REQUIRED
	}

	// Category is the show/episodes category and it's subcategories
	Category struct {
		Name        string   `json:"name" yaml:"name" binding:"required"`      // REQUIRED
		SubCategory []string `json:"subcategory" yaml:"subcategory,omitempty"` // OPTIONAL
	}

	AssetRef struct {
		URI       string `json:"uri" yaml:"uri" binding:"required"`              // REQUIRED
		Rel       string `json:"rel" yaml:"rel" binding:"required"`              // REQUIRED
		Type      string `json:"type,omitempty" yaml:"type,omitempty"`           // OPTIONAL
		ETag      string `json:"etag,omitempty" yaml:"etag,omitempty"`           // OPTIONAL
		Duration  int    `json:"duration,omitempty" yaml:"duration,omitempty"`   // OPTIONAL
		Timestamp int64  `json:"timestamp,omitempty" yaml:"timestamp,omitempty"` // OPTIONAL
		Size      int    `json:"size,omitempty" yaml:"size,omitempty"`           // OPTIONAL
	}
)

// GUID is a convenience method to access the resources guid
func (r *GenericResource) GUID() string {
	return r.Metadata.GUID
}

// GUID is a convenience method to access the resources guid
func (s *Show) GUID() string {
	return s.Metadata.GUID
}

// PublishDateTimestamp converts a RFC1123Z formatted timestamp into UNIX timestamp
func (e *Episode) PublishDateTimestamp() int64 {
	pd := e.Metadata.Date
	if pd == "" {
		return 0
	}
	t, err := time.Parse(time.RFC1123Z, pd)
	if err != nil {
		return 0
	}

	return t.Unix()
}

// PublishDate is a convenience method to access the pub date
func (e *Episode) PublishDate() string {
	return e.Metadata.Date
}

// GUID is a convenience method to access the resources guid
func (e *Episode) GUID() string {
	return e.Metadata.GUID
}

// ParentGUID is a convenience method to access the resources parent guid
func (e *Episode) Parent() string {
	return e.Metadata.Parent
}

// LocalNamePart returns the part after the last /, if any
func (r *AssetRef) LocalNamePart() string {
	parts := strings.Split(r.URI, "/")
	return parts[len(parts)-1:][0]
}

// DefaultShowMetadata creates a default set of labels etc for a Show resource
//
//	language:	<ISO639 two-letter-code> REQUIRED 'channel.language'
//	explicit:	True | False REQUIRED 'channel.itunes.explicit'
//	type:		Episodic | Serial REQUIRED 'channel. itunes.type'
//	block:		Yes OPTIONAL 'channel.itunes.block' Anything else than 'Yes' has no effect
//	complete:	Yes OPTIONAL 'channel.itunes.complete' Anything else than 'Yes' has no effect
func DefaultShowMetadata() map[string]string {
	l := make(map[string]string)

	l[LabelLanguage] = "en_US"
	l[LabelExplicit] = "no"
	l[LabelType] = ShowTypeEpisodic
	l[LabelBlock] = "no"
	l[LabelComplete] = "no"

	return l
}

// DefaultEpisodeMetadata creates a default set of labels etc for a Episode resource
//	season: 	<season number> OPTIONAL 'item.itunes.season'
//	episode:	<episode number> REQUIRED 'item.itunes.episode'
//	explicit:	True | False REQUIRED 'channel.itunes.explicit'
//	type:		Full | Trailer | Bonus REQUIRED 'item.itunes.episodeType'
//	block:		Yes OPTIONAL 'item.itunes.block' Anything else than 'Yes' has no effect
func DefaultEpisodeMetadata() map[string]string {
	l := make(map[string]string)

	l[LabelSeason] = "1"
	l[LabelEpisode] = "1"
	l[LabelExplicit] = "no"
	l[LabelType] = EpisodeTypeFull
	l[LabelBlock] = "no"

	return l
}

// DefaultShow creates a default show struc
func DefaultShow(name, title, summary, guid, portal, cdn string) *Show {
	return &Show{
		APIVersion: config.Version,
		Kind:       ResourceShow,
		Metadata: Metadata{
			Name:   name,
			GUID:   guid,
			Date:   time.Now().UTC().Format(time.RFC1123Z),
			Labels: DefaultShowMetadata(),
		},
		Description: ShowDescription{
			Title:   title,
			Summary: summary,
			Link: AssetRef{
				URI: fmt.Sprintf("%s/%s", portal, name),
				Rel: ResourceTypeExternal,
			},
			Category: []Category{
				{
					Name: "Technology",
					SubCategory: []string{
						"Podcasting",
					},
				}},
			Owner: Owner{
				Name:  "PODCAST OWNER(S)",
				Email: "HELLO@PODCAST",
			},
			Author:    "PODCAST AUTHOR(S)",
			Copyright: "PODCAST COPYRIGHT",
		},
		Image: AssetRef{
			URI: "cover.png",
			Rel: ResourceTypeLocal,
		},
	}
}

// DefaultEpisode creates a default episode struc
func DefaultEpisode(name, parentName, guid, parent, portal, cdn string) *Episode {
	return &Episode{
		APIVersion: config.Version,
		Kind:       ResourceEpisode,
		Metadata: Metadata{
			Name:   name,
			GUID:   guid,
			Parent: parent,
			Date:   time.Now().UTC().Format(time.RFC1123Z),
			Labels: DefaultEpisodeMetadata(),
		},
		Description: EpisodeDescription{
			Title:       "EPISODE TITLE",
			Summary:     "EPISODE SUMMARY",
			EpisodeText: "EPISODE DESCRIPTION",
			Link: AssetRef{
				URI: fmt.Sprintf("%s/%s/%s", portal, parentName, name),
				Rel: ResourceTypeExternal,
			},
			Duration: 1, // Seconds. Must not be 0, otherwise a validation error occurs.
		},
		Image: AssetRef{
			URI: "episode.png",
			Rel: ResourceTypeLocal,
		},
		Enclosure: AssetRef{
			URI:  "episode.mp3",
			Type: "audio/mpeg",
			Rel:  ResourceTypeLocal,
		},
	}
}
