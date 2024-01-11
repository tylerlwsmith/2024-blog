package wp

import (
	"net/url"
	"time"
)

type WPPost struct {
	Id            int
	Date          time.Time
	DateGmt       time.Time
	Guid          WPGuid
	Modified      time.Time
	ModifiedGmt   time.Time
	Slug          string
	Type          string
	Link          url.URL
	Title         WPRenderable
	Content       WPRenderable
	Excerpt       WPRenderable
	Author        int
	FeaturedMedia int
	CommentStatus string
	PingStatus    string
	Sticky        bool
	Template      string
	Format        string
	Meta          WPMeta
	Categories    []int
	Tags          []int
}

type WPGuid struct {
	Rendered url.URL
}

type WPRenderable struct {
	Rendered  string
	Protected bool
}

type WPMeta struct {
	Footnotes string
}
