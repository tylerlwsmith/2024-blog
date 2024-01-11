package wp

import (
	"net/url"
	"time"
)

type WPPost struct {
	Id            int          `json:"id"`
	Date          time.Time    `json:"date"`
	DateGmt       time.Time    `json:"date_gmt"`
	Guid          WPGuid       `json:"guid"`
	Modified      time.Time    `json:"modified"`
	ModifiedGmt   time.Time    `json:"modified_gmt"`
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
