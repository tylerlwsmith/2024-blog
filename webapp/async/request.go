package async

import (
	"net/http"
	"sync"
)

type requestResult struct {
	res *http.Response
	err error
}

type AsyncRequest struct {
	channel chan *requestResult
	result  *requestResult
	once    *sync.Once
}

func (r *AsyncRequest) AwaitResponse() (resp *http.Response, err error) {
	r.once.Do(func() {
		r.result = <-r.channel
	})

	return r.result.res, r.result.err
}

func newAsyncRequest(c chan *requestResult) AsyncRequest {
	return AsyncRequest{
		channel: c,
		once:    &sync.Once{},
	}
}

func (r requestResult) Result() (res *http.Response, err error) {
	return r.res, r.err
}

func Get(url string) AsyncRequest {
	c := make(chan *requestResult)
	go func(url string) {
		resp, err := http.Get(url)
		c <- &requestResult{resp, err}
	}(url)

	return newAsyncRequest(c)
}
