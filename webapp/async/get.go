package async

import (
	"net/http"
	"sync"
)

type response struct {
	res *http.Response
	err error
}

type ResponsePromise struct {
	channel chan *response
	result  *response
	once    *sync.Once
}

func (r *ResponsePromise) AwaitResponse() (resp *http.Response, err error) {
	r.once.Do(func() {
		r.result = <-r.channel
	})

	return r.result.res, r.result.err
}

func Get(url string) ResponsePromise {
	c := make(chan *response)
	go func(url string) {
		resp, err := http.Get(url)
		c <- &response{resp, err}
	}(url)

	return newResponsePromise(c)
}

func newResponsePromise(c chan *response) ResponsePromise {
	return ResponsePromise{
		channel: c,
		once:    &sync.Once{},
	}
}
