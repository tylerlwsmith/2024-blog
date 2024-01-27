package wp

import (
	"encoding/json"
	"html/template"
	"net/url"
	"strings"
	"time"
)

type WPTag struct {
	Id          int    `json:"id"`
	Count       int    `json:"count"`
	Description string `json:"description"`
	Link        WPURL  `json:"link"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	// There's an excluded "meta" field because I don't know what's in it.
}

func (t WPTag) HasDescription() bool {
	return strings.TrimSpace(string(t.Description)) != ""
}

type WPPost struct {
	Id            int          `json:"id"`
	Date          WPTime       `json:"date"`
	DateGmt       WPTime       `json:"date_gmt"`
	Guid          WPGuid       `json:"guid"`
	Modified      WPTime       `json:"modified"`
	ModifiedGmt   WPTime       `json:"modified_gmt"`
	Slug          string       `json:"slug"`
	Type          string       `json:"type"`
	Link          WPURL        `json:"link"`
	Title         WPRenderable `json:"title"`
	Content       WPRenderable `json:"content"`
	Excerpt       WPRenderable `json:"excerpt"`
	Author        int          `json:"author"`
	FeaturedMedia int          `json:"featured_media"`
	CommentStatus string       `json:"comment_status"`
	PingStatus    string       `json:"ping_status"`
	Sticky        bool         `json:"sticky"`
	Template      string       `json:"template"`
	Format        string       `json:"format"`
	Meta          WPMeta       `json:"meta"`
	Categories    []int        `json:"categories"`
	Tags          []int        `json:"tags"`
}

func (p WPPost) HasContent() bool {
	return strings.TrimSpace(string(p.Content.Rendered)) != ""
}

type WPTime struct {
	time.Time
}

// https://core.trac.wordpress.org/ticket/51945
// https://eli.thegreenplace.net/2020/unmarshaling-time-values-from-json/
// https://ukiahsmith.com/blog/improved-golang-unmarshal-json-with-time-and-url/
func (t *WPTime) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	RFC3339WithoutTimezone := "2006-01-02T15:04:05"
	t.Time, err = time.Parse(RFC3339WithoutTimezone, s)
	return err
}

type WPURL struct {
	url.URL
}

func (u *WPURL) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := url.Parse(s)
	if err != nil {
		return err
	}

	u.URL = *parsed
	return nil
}

type WPGuid struct {
	Rendered WPURL `json:"rendered"`
}

type WPRenderable struct {
	Rendered  template.HTML `json:"rendered"`
	Protected bool          `json:"protected,omitempty"`
}

type WPMeta struct {
	Footnotes string `json:"footnotes"`
}

type WPUser struct {
	Id                int               `json:"id"`
	Username          string            `json:"username"`
	Name              string            `json:"name"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	Email             string            `json:"email"`
	URL               WPURL             `json:"url"`
	Description       string            `json:"description"`
	Link              WPURL             `json:"link"`
	Locale            string            `json:"locale"`
	Nickname          string            `json:"nickname"`
	Slug              string            `json:"slug"`
	Roles             []string          `json:"roles"`
	RegisteredDate    string            `json:"registered_date"` // Should be some kind of time object
	Capabilities      map[string]bool   `json:"capabilities"`
	ExtraCapabilities map[string]bool   `json:"extra_capabilities"`
	AvatarURLs        map[string]string `json:"avatar_urls"`
}

func (u WPUser) CanEditPost(p WPPost) bool {
	if !u.Exists() {
		return false
	}
	if val, ok := u.ExtraCapabilities["administrator"]; ok && val {
		return true
	}
	if val, ok := u.Capabilities["edit_others_posts"]; ok && val {
		return true
	}
	if val, ok := u.Capabilities["edit_posts"]; ok && val && u.Id == p.Author {
		return true
	}
	return false
}

func (u WPUser) CanEditTag(p WPTag) bool {
	if !u.Exists() {
		return false
	}
	if val, ok := u.ExtraCapabilities["administrator"]; ok && val {
		return true
	}
	if val, ok := u.Capabilities["manage_categories"]; ok && val {
		return true
	}
	return false
}

func (u WPUser) Exists() bool {
	return u.Id != 0
}

type WPNonce struct {
	Nonce string `json:"nonce"`
}
