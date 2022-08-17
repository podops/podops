package podops

import (
	"fmt"
	"time"

	"github.com/podops/podops/config"
)

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
	l[LabelExplicit] = "False"
	l[LabelType] = ShowTypeEpisodic
	l[LabelBlock] = "No"
	l[LabelComplete] = "No"

	return l
}

// DefaultEpisodeMetadata creates a default set of labels etc for a Episode resource
//
//	season: 	<season number> OPTIONAL 'item.itunes.season'
//	episode:	<episode number> REQUIRED 'item.itunes.episode'
//	explicit:	True | False REQUIRED 'channel.itunes.explicit'
//	type:		Full | Trailer | Bonus REQUIRED 'item.itunes.episodeType'
//	block:		Yes OPTIONAL 'item.itunes.block' Anything else than 'Yes' has no effect
func DefaultEpisodeMetadata() map[string]string {
	l := make(map[string]string)

	l[LabelSeason] = "1"
	l[LabelEpisode] = "1"
	l[LabelExplicit] = "False"
	l[LabelType] = EpisodeTypeFull
	l[LabelBlock] = "No"

	return l
}
