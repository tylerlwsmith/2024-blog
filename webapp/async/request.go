package async

import (
	"net/http"
)

type RequestResult struct {
	res *http.Response
	err error
}

func (r RequestResult) Result() (res *http.Response, err error) {
	return r.res, r.err
}

func Get(url string) chan *RequestResult {
	c := make(chan *RequestResult)
	go func(url string) {
		resp, err := http.Get(url)
		c <- &RequestResult{resp, err}
	}(url)

	return c
}
