package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"webapp/wp"
)

func unmarshalAPIRequest[T any](url string, value *T) (err error) {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return errors.New("API returned non-200 status code")
	}

	bytes, err := io.ReadAll(res.Body)

	err = json.Unmarshal(bytes, &value)

	return err
}

func GetPosts() (posts []wp.WPPost, err error) {
	posts = []wp.WPPost{}
	err = unmarshalAPIRequest[[]wp.WPPost]("http://wordpress:80/wp-json/wp/v2/posts", &posts)
	return posts, err
}

func GetTags() (tags []wp.WPTag, err error) {
	tags = []wp.WPTag{}
	err = unmarshalAPIRequest[[]wp.WPTag]("http://wordpress:80/wp-json/wp/v2/posts", &tags)
	return tags, err
}
