package builder

import (
	"bytes"
	"time"

	"github.com/yuin/goldmark"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/rss"
)

// transformToPodcast transforms Show metadata into a podcast feed struct
func transformToPodcast(s *podops.Show) (*rss.Channel, error) {
	now := time.Now()

	// basics
	pf := rss.New(s.Description.Title, s.Description.Link.URI, &now, &now)

	// details
	summary, err := renderMarkdown(s.Description.Summary)
	if err != nil {
		summary = s.Description.Summary
	}
	pf.AddSummary(summary)
	if s.Description.Author == "" {
		pf.AddAuthor(s.Description.Owner.Name, s.Description.Owner.Email)
	} else {
		pf.IAuthor = s.Description.Author
	}

	// add the podcast category. see https://help.apple.com/itc/podcasts_connect/#/itc9267a2f12
	for _, category := range s.Description.Category {
		pf.AddCategory(category.Name, category.SubCategory)
	}

	pf.AddImage(s.Image.CanonicalReference(config.Settings().GetOption(config.PodopsContentEndpointEnv), s.GUID()))
	pf.IOwner = &rss.Author{
		Name:  s.Description.Owner.Name,
		Email: s.Description.Owner.Email,
	}
	pf.Copyright = s.Description.Copyright
	if s.NewFeedLink != nil {
		pf.INewFeedURL = s.NewFeedLink.URI
	}
	pf.Language = s.Metadata.Labels[podops.LabelLanguage]
	pf.IExplicit = s.Metadata.Labels[podops.LabelExplicit]

	t := s.Metadata.Labels[podops.LabelType]
	if t == podops.ShowTypeEpisodic || t == podops.ShowTypeSerial {
		pf.IType = t
	} else {
		return nil, podops.ErrInvalidParameters
	}
	if s.Metadata.Labels[podops.LabelBlock] == "yes" {
		pf.IBlock = "yes"
	}
	if s.Metadata.Labels[podops.LabelComplete] == "yes" {
		pf.IComplete = "yes"
	}

	return &pf, nil
}

//	explicit:	True | False REQUIRED 'channel.itunes.explicit'
//	type:		Episodic | Serial REQUIRED 'channel. itunes.type'
//	block:		Yes OPTIONAL 'channel.itunes.block' Anything else than 'Yes' has no effect
//	complete:	Yes OPTIONAL 'channel.itunes.complete' Anything else than 'Yes' has no effect

// transformToItem returns the episode struct needed for a podcast feed struct
func transformToItem(e *podops.Episode) (*rss.Item, error) {

	pubDate, err := time.Parse(time.RFC1123Z, e.Metadata.Date) // FIXME this is redunant
	if err != nil {
		return nil, err
	}

	ef := &rss.Item{
		Title:       e.Description.Title,
		Description: e.Description.Summary,
	}

	ef.AddEnclosure(e.Enclosure.CanonicalReference(config.Settings().GetOption(config.PodopsContentEndpointEnv), e.Parent()), mediaTypeMap[e.Enclosure.Type], (int64)(e.Enclosure.Size))
	ef.AddImage(e.Image.CanonicalReference(config.Settings().GetOption(config.PodopsContentEndpointEnv), e.Parent()))
	ef.AddPubDate(&pubDate)
	ef.AddSummary(e.Description.EpisodeText)

	if e.Enclosure.Duration > 1 {
		ef.AddDuration((int64)(e.Enclosure.Duration))
	} else {
		ef.AddDuration((int64)(e.Description.Duration))
	}

	ef.Link = e.Description.Link.CanonicalReference(config.Settings().GetOption(config.PodopsContentEndpointEnv), e.Parent())
	ef.ISubtitle = e.Description.Summary
	ef.GUID = e.Metadata.GUID
	ef.IExplicit = e.Metadata.Labels[podops.LabelExplicit]
	ef.ISeason = e.Metadata.Labels[podops.LabelSeason]
	ef.IEpisode = e.Metadata.Labels[podops.LabelEpisode]
	ef.IEpisodeType = e.Metadata.Labels[podops.LabelType]
	if e.Metadata.Labels[podops.LabelBlock] == "yes" {
		ef.IBlock = "yes"
	}

	return ef, nil
}

func renderMarkdown(md string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
