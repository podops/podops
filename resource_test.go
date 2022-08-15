package podops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/txsvc/stdlib/v2/validate"
)

const (
	testPodcastName       string = "simple-podcast"
	testPodcastTitle      string = "Simple PodOps SDK Example"
	testPodcastSummary    string = "A simple podcast for testing and experimentation. Created with the PodOps API."
	testPodcastGUID       string = "c8d34e24b230"
	testPodcastParentGUID string = "2ffc49bda824"
)

func TestValidateName(t *testing.T) {
	assert.True(t, ValidName("this_is-valid"))
	assert.False(t, ValidName("short"))
	assert.False(t, ValidName("this_is_too_long-this_is_too_long-this_is_too_long_too_long_too_long"))
	assert.False(t, ValidName("no spaces_allowed"))
}

func TestValidateEmail(t *testing.T) {
	assert.True(t, ValidEmail("me@example.com"))
	assert.True(t, ValidEmail("me@example")) // TLD missing but that's OK
	// FIXME validate the tests
	assert.False(t, ValidEmail("me"))         // too short
	assert.False(t, ValidEmail("meexample"))  // @ missing
	assert.False(t, ValidEmail("me example")) // no whitespace allowed
}

func TestDefaultShowMetadata(t *testing.T) {
	meta := DefaultShowMetadata()
	assert.NotEmpty(t, meta)
	assert.Equal(t, 5, len(meta))

	assert.NotEmpty(t, meta[LabelLanguage])
	assert.NotEmpty(t, meta[LabelExplicit])
	assert.NotEmpty(t, meta[LabelType])
	assert.NotEmpty(t, meta[LabelBlock])
	assert.NotEmpty(t, meta[LabelComplete])
}

func TestDefaultEpisodeMetadata(t *testing.T) {
	meta := DefaultEpisodeMetadata()
	assert.NotEmpty(t, meta)
	assert.Equal(t, 5, len(meta))

	assert.NotEmpty(t, meta[LabelSeason])
	assert.NotEmpty(t, meta[LabelEpisode])
	assert.NotEmpty(t, meta[LabelExplicit])
	assert.NotEmpty(t, meta[LabelType])
	assert.NotEmpty(t, meta[LabelBlock])
}

func TestDefaultShowResource(t *testing.T) {
	show := DefaultShow(testPodcastName, testPodcastTitle, testPodcastSummary, testPodcastGUID, "portal", "cdn")
	assert.NotNil(t, show)

	v := show.Validate("show_test", validate.NewValidator())
	assert.Equal(t, 0, v.Errors)

	assert.Equal(t, testPodcastGUID, show.GUID())
}

func TestDefaultEpisodeResource(t *testing.T) {
	episode := DefaultEpisode(testPodcastName, "parent_name", testPodcastGUID, testPodcastParentGUID, "portal", "cdn")
	assert.NotNil(t, episode)

	v := episode.Validate("episode_test", validate.NewValidator())
	assert.Equal(t, 0, v.Errors)

	assert.Equal(t, testPodcastGUID, episode.GUID())
	assert.Equal(t, testPodcastParentGUID, episode.Parent())
	assert.NotEmpty(t, episode.PublishDate())
	assert.Greater(t, episode.PublishDateTimestamp(), int64(0))
}

func TestNormalizeKind(t *testing.T) {
	assert.NotEmpty(t, resourceMap)

	alias, err := NormalizeKind("foobar")
	assert.Error(t, err)
	assert.Empty(t, alias)

	alias, err = NormalizeKind("shows")
	assert.NoError(t, err)
	assert.Equal(t, alias, "show")

	alias, err = NormalizeKind("episodes")
	assert.NoError(t, err)
	assert.Equal(t, alias, "episode")

	alias, err = NormalizeKind("assets")
	assert.NoError(t, err)
	assert.Equal(t, alias, "asset")

	alias, err = NormalizeKind("all")
	assert.NoError(t, err)
	assert.Equal(t, alias, "all")
}

func TestShowResource(t *testing.T) {
	show := DefaultShow(testPodcastName, testPodcastTitle, testPodcastSummary, testPodcastGUID, "portal", "cdn")
	assert.NotNil(t, show)
	assert.Equal(t, testPodcastGUID, show.GUID())

	show.Metadata.Name = ""

	v := show.Validate("show_test", validate.NewValidator())
	assert.Equal(t, 1, v.Errors)
}
