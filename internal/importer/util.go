package importer

import (
	"strings"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"

	"github.com/podops/podops"
	"github.com/podops/podops/internal"
)

func convDurationToInt(ite *ext.ITunesItemExtension) int {
	if ite == nil {
		return 0
	}
	return internal.ConvTimeStringToSeconds(ite.Duration)
}

func itemGUID(item *gofeed.Item) string {
	if item.GUID != "" {
		if podops.ValidGUID(item.GUID) {
			return item.GUID
		}
		return internal.CreateShortGUID(item.GUID)
	}
	return internal.CreateShortGUID(item.Link)
}

func stringWithDefault(s, def string) string {
	if s == "" {
		return def
	}
	return strings.Trim(s, " ")
}

// stringExpect compares s and exp, all case insensitive. If a match, exp is returned, def otherwise.
func stringExpect(s, exp, def string) string {
	if s == "" {
		return def
	}
	if strings.EqualFold(s, exp) {
		return exp
	}
	return def
}

func formatName(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.Trim(s, " "), " ", "_"))
}
