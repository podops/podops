package importer

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	urls = []string{
		"https://changelog.com/gotime/feed",
		"https://trojaalert.bildungsangst.de/feed/mp3/",
		"https://changelog.com/master/feed",
		"https://feeds.megaphone.fm/goforbroke",
		"https://feeds.feedburner.com/afterhours_tac",
		"https://podcasts.mckinsey.com/fp/futureofasia_itunes",
		"https://feeds.buzzsprout.com/1004689.rss",
		"https://kackundsach.podigee.io/feed/mp3",
		"https://feeds.blubrry.com/feeds/microsoftresearch.xml",
		"https://deloitteus.libsyn.com/rss",
		"https://feeds.feedburner.com/TedInterview",

		"http://feeds.hoaxilla.com/hoaxilla",
		"https://resonator-podcast.de/?feed=m4a",
		"http://chaosradio.ccc.de/chaosradio-latest.rss",
		"http://alternativlos.org/alternativlos.rss",
		"https://feeds.metaebene.me/nsfw/m4a",
		"https://ukw.fm/feed/mp3/",
		"https://feeds.metaebene.me/forschergeist/m4a",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestRetrieveFeed(t *testing.T) {
	// select a random url
	url := urls[rand.Intn(len(urls)-1)]
	fmt.Printf("Retrieving: %s\n", url)

	show, err := ImportPodcastFeed(url)

	assert.NoError(t, err)
	assert.NotNil(t, show)
}
