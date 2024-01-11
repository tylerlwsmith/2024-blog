package wp

import (
	"net/url"
	"strings"
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
	Link          url.URL      `json:"url"`
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
func (t *WPTime) UnmarshalJSON(data []byte) (err error) {
	stringDate := strings.Trim(string(data), "\"")

	var nt time.Time
	err = nt.UnmarshalJSON([]byte("\"" + stringDate + "Z\""))
	t.Time = nt
	return err
}

type WPGuid struct {
	Rendered url.URL `json:"rendered"`
}

type WPRenderable struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected,omitempty"`
}

type WPMeta struct {
	Footnotes string `json:"footnotes"`
}
