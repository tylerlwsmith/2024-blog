package wp

import (
	"encoding/json"
	"net/url"
	"time"
)

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
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected,omitempty"`
}

type WPMeta struct {
	Footnotes string `json:"footnotes"`
}
