package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"webapp/wp"
)

type apiRequest[T any] struct {
	endpoint string
	query    map[string]any
}

func newRequest[T any](url string) *apiRequest[T] {
	return &apiRequest[T]{endpoint: url}
}

// TODO: this doesn't work yet.
func (req *apiRequest[T]) SetParam(key string, value any) *apiRequest[T] {
	req.query[key] = value
	return req
}

// TODO: this doesn't work yet.
func (req *apiRequest[T]) First() (value T, header http.Header, err error) {
	var values []T
	header, err = unmarshalAPIRequest[[]T](req.endpoint, &values)
	// todo: rip first value out, return error if no values.
	return value, header, err
}

func (req *apiRequest[T]) Get() (values []T, header http.Header, err error) {
	header, err = unmarshalAPIRequest[[]T](req.endpoint, &values)
	return values, header, err
}

// TODO: GetAll() method that loops through all pages.

func unmarshalAPIRequest[T any](url string, value *T) (header http.Header, err error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return res.Header, errors.New("API returned non-200 status code")
	}

	bytes, err := io.ReadAll(res.Body)

	err = json.Unmarshal(bytes, &value)

	return res.Header, err
}

func Posts() (request *apiRequest[wp.WPPost]) {
	return newRequest[wp.WPPost]("http://wordpress:80/wp-json/wp/v2/posts")
}

func Tags() (request *apiRequest[wp.WPTag]) {
	return newRequest[wp.WPTag]("http://wordpress:80/wp-json/wp/v2/tags")
}
