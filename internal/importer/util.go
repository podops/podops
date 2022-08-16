package importer

import (
	"strings"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/podops/podops/internal"
	"github.com/txsvc/stdlib/v2/id"
)

func convDurationToInt(ite *ext.ITunesItemExtension) int {
	if ite == nil {
		return 0
	}
	return internal.ConvTimeStringToSeconds(ite.Duration)
}

func itemGUID(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	return id.Fingerprint(item.Link)
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
