package podops

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/txsvc/stdlib/v2/validate"

	"github.com/podops/podops/config"
)

var (
	nameRegex = regexp.MustCompile(`^[a-z]+[a-z0-9_-]`)
	guidRegex = regexp.MustCompile(`^[a-f0-9]`)
	// FIXME validate the regex for email
	//emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	// mapping of resource names and aliases
	resourceMap map[string]string
)

func init() {
	resourceMap = make(map[string]string)
	resourceMap["show"] = "show"
	resourceMap["shows"] = "show"
	resourceMap["episode"] = "episode"
	resourceMap["episodes"] = "episode"
	resourceMap["asset"] = "asset"
	resourceMap["assets"] = "asset"
	resourceMap["all"] = "all"
}

func NormalizeKind(kind string) (string, error) {
	if k, ok := resourceMap[strings.ToLower(kind)]; ok {
		return k, nil
	}
	return "", fmt.Errorf(MsgResourceIsInvalid, kind)
}

// ValidName verifies that a name is valid for a resource. The following rules apply:
//
// 'name' must contain only lowercase letters, numbers, dashes (-), underscores (_).
// 'name' must contain 8-64 characters.
// Spaces and dots (.) are not allowed.
func ValidName(name string) bool {
	if len(name) < 8 || len(name) > 64 {
		return false
	}
	if strings.Contains(name, " ") {
		return false
	}
	return nameRegex.MatchString(name)
}

func ValidGUID(guid string) bool {
	if len(guid) != 12 {
		return false
	}
	if strings.Contains(guid, " ") {
		return false
	}
	return guidRegex.MatchString(guid)
}

// ValidEmail checks if the email provided passes the required structure and length.
func ValidEmail(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

// Validate verifies the integrity of struct Show
//
//	APIVersion  string          `json:"apiVersion" yaml:"apiVersion" binding:"required"`   // REQUIRED default: v1.0
//	Kind        string          `json:"kind" yaml:"kind" binding:"required"`               // REQUIRED default: show
//	Metadata    Metadata        `json:"metadata" yaml:"metadata" binding:"required"`       // REQUIRED
//	Description ShowDescription `json:"description" yaml:"description" binding:"required"` // REQUIRED
//	Image       Resource        `json:"image" yaml:"image" binding:"required"`             // REQUIRED 'channel.itunes.image'
func (s *Show) Validate(root string, v *validate.Validator) *validate.Validator {
	v.SaveContext("show")
	defer v.RestoreContext()

	v.StringEquals(s.APIVersion, config.Version)
	v.StringEquals(s.Kind, ResourceShow)

	// Show specific metadata, tracking the scaffolding functions
	s.Metadata.Validate(root+".metadata", v)

	v.ISO639(s.Metadata.Labels[LabelLanguage])
	v.MapContains(s.Metadata.Labels, LabelLanguage, "Metadata")
	v.MapContains(s.Metadata.Labels, LabelExplicit, "Metadata")
	v.MapContains(s.Metadata.Labels, LabelType, "Metadata")
	v.MapContains(s.Metadata.Labels, LabelBlock, "Metadata")
	v.MapContains(s.Metadata.Labels, LabelComplete, "Metadata")

	s.Description.Validate(root+".description", v)
	s.Image.Validate(root+".image", v)

	return v
}

// Validate verifies the integrity of struct Episode
//
//	APIVersion  string             `json:"apiVersion" yaml:"apiVersion" binding:"required"`   // REQUIRED default: v1.0
//	Kind        string             `json:"kind" yaml:"kind" binding:"required"`               // REQUIRED default: episode
//	Metadata    Metadata           `json:"metadata" yaml:"metadata" binding:"required"`       // REQUIRED
//	Description EpisodeDescription `json:"description" yaml:"description" binding:"required"` // REQUIRED
//	Image       Resource           `json:"image" yaml:"image" binding:"required"`             // REQUIRED 'item.itunes.image'
//	Enclosure   Resource           `json:"enclosure" yaml:"enclosure" binding:"required"`     // REQUIRED
func (e *Episode) Validate(root string, v *validate.Validator) *validate.Validator {
	v.SaveContext("episode")
	defer v.RestoreContext()

	v.StringEquals(e.APIVersion, config.Version)
	v.StringEquals(e.Kind, ResourceEpisode)

	// Episode specific metadata, tracking the scaffolding functions
	e.Metadata.Validate(root+".metadata", v)

	v.MapContains(e.Metadata.Labels, LabelSeason, "Metadata")
	v.MapContains(e.Metadata.Labels, LabelEpisode, "Metadata")
	v.MapContains(e.Metadata.Labels, LabelExplicit, "Metadata")
	v.MapContains(e.Metadata.Labels, LabelType, "Metadata")
	v.MapContains(e.Metadata.Labels, LabelBlock, "Metadata")

	e.Description.Validate(root+".description", v)
	e.Image.Validate(root+".image", v)
	e.Enclosure.Validate(root+".enclosure", v)

	return v
}

// Validate verifies the integrity of struct Metadata
//
//	Name   string            `json:"name" yaml:"name" binding:"required"` // REQUIRED
//  GUID   string            `json:"guid" yaml:"guid" binding:"required"` // REQUIRED
func (m *Metadata) Validate(root string, v *validate.Validator) *validate.Validator {
	v.SaveContext(root)
	defer v.RestoreContext()

	if !ValidName(m.Name) {
		v.AddError(fmt.Sprintf(MsgResourceInvalidName, m.Name))
	}
	if !ValidGUID(m.GUID) {
		v.AddError(fmt.Sprintf(MsgResourceInvalidGUID, m.Name, m.GUID))
	}

	if m.Parent != "" {
		if !ValidGUID(m.Parent) {
			v.AddError(fmt.Sprintf(MsgResourceInvalidGUID, m.Name, m.Parent))
		}
	}

	return v
}

// Validate verifies the integrity of struct ShowDescription
//
//	Title     string    `json:"title" yaml:"title" binding:"required"`          // REQUIRED 'channel.title' 'channel.itunes.title'
//	Summary   string    `json:"summary" yaml:"summary" binding:"required"`      // REQUIRED 'channel.description'
//	Link      Resource  `json:"link" yaml:"link"`                               // RECOMMENDED 'channel.link'
//	Category  Category  `json:"category" yaml:"category" binding:"required"`    // REQUIRED channel.category
//	Owner     Owner     `json:"owner" yaml:"owner"`                             // RECOMMENDED 'channel.itunes.owner'
//	Author    string    `json:"author" yaml:"author"`                           // RECOMMENDED 'channel.itunes.author'
//	Copyright string    `json:"copyright,omitempty" yaml:"copyright,omitempty"` // OPTIONAL 'channel.copyright'
//	NewFeed   *Resource `json:"newFeed,omitempty" yaml:"newFeed,omitempty"`     // OPTIONAL channel.itunes.new-feed-url -> move to label
func (d *ShowDescription) Validate(root string, v *validate.Validator) *validate.Validator {
	v.SaveContext(root)
	defer v.RestoreContext()

	v.StringNotEmpty(d.Title, "Title")
	v.StringNotEmpty(d.Summary, "Summary")

	d.Link.Validate(root, v)
	d.Owner.Validate(root, v)

	if len(d.Category) > 0 {
		for _, c := range d.Category {
			c.Validate(root, v)
		}
	} else {
		v.AddError(MsgMissingCategory)
	}

	return v
}

// Validate verifies the integrity of struct EpisodeDescription
//
//	Title       string   `json:"title" yaml:"title" binding:"required"`                                 // REQUIRED 'item.title' 'item.itunes.title'
//	Summary     string   `json:"summary" yaml:"summary" binding:"required"`                             // REQUIRED 'item.description'
//	EpisodeText string   `json:"episodeText,omitempty" yaml:"episodeText,omitempty" binding:"required"` // REQUIRED 'item.itunes.summary'
//	Link        Resource `json:"link" yaml:"link"`                                                      // RECOMMENDED 'item.link'
//	Duration    int      `json:"duration" yaml:"duration" binding:"required"`                           // REQUIRED 'item.itunes.duration'
func (d *EpisodeDescription) Validate(root string, v *validate.Validator) *validate.Validator {
	v.SaveContext(root)
	defer v.RestoreContext()

	v.StringNotEmpty(d.Title, "Title")
	v.StringNotEmpty(d.Summary, "Summary")
	v.StringNotEmpty(d.EpisodeText, "EpisodeText")

	d.Link.Validate(root, v)

	v.NonZero(d.Duration, "Duration")

	return v
}

// Validate verifies the integrity of struct Resource
//
//	URI    string `json:"uri" yaml:"uri" binding:"required"`        // REQUIRED
//	Rel    string `json:"rel,omitempty" yaml:"rel,omitempty"`       // REQUIRED
//	Type   string `json:"type,omitempty" yaml:"type,omitempty"`     // OPTIONAL
//	Size   int    `json:"size,omitempty" yaml:"size,omitempty"`     // OPTIONAL
func (r *AssetRef) Validate(root string, v *validate.Validator) *validate.Validator {
	v.SaveContext(root)
	defer v.RestoreContext()

	v.StringNotEmpty(r.URI, "URI")
	if !validate.IsMemberOf(r.Rel, ResourceTypeLocal, ResourceTypeExternal, ResourceTypeImport) {
		v.AddError(fmt.Sprintf(MsgResourceInvalidReference, r.Rel))
	}
	return v
}

// Validate verifies the integrity of struct Category
//
//	Name        string   `json:"name" yaml:"name" binding:"required"`      // REQUIRED
//	SubCategory []string `json:"subcategory" yaml:"subcategory,omitempty"` // OPTIONAL
func (c *Category) Validate(root string, v *validate.Validator) *validate.Validator {
	v.StringNotEmpty(c.Name, "Name")
	return v
}

// Validate verifies the integrity of struct Owner
//
//	Name  string `json:"name" yaml:"name" binding:"required"`   // REQUIRED
//	Email string `json:"email" yaml:"email" binding:"required"` // REQUIRED
func (o *Owner) Validate(root string, v *validate.Validator) *validate.Validator {
	v.StringNotEmpty(o.Name, "Name")
	v.StringNotEmpty(o.Email, "EMail")
	if !ValidEmail(o.Email) {
		v.AddError(fmt.Sprintf(MsgInvalidEmail, o.Email))
	}
	return v
}
