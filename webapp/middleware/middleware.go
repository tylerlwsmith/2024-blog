package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"webapp/wp"
)

func StripTrailingSlashesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path != "/" && path[len(path)-1:] == "/" {
			for path[len(path)-1:] == "/" {
				path = path[:len(path)-1]
			}

			// Clone URL. I'm ~20% sure this is doing what I think it's doing.
			// I'm also not absolutely certain that this needs to be cloned.
			// https://github.com/golang/go/issues/41733#issuecomment-708556495
			c := *r.URL
			c.Path = path

			http.Redirect(w, r, path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserMiddleware(next http.Handler) http.Handler {
	// http://localhost:8080/wp-json/wp/v2/users/me?context=edit&_wpnonce=somevalue
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u wp.WPUser
		client := http.Client{}
		func() {
			nReq, err := http.NewRequest("GET", "http://wordpress/wp-json/nonce/v1/nonce", nil)
			if err != nil {
				return
			}

			for _, cookie := range r.Cookies() {
				nReq.AddCookie(cookie)
			}

			nRes, err := client.Do(nReq)
			if err != nil || nRes.StatusCode != 200 {
				return
			}
			defer nRes.Body.Close()

			nBytes, err := io.ReadAll(nRes.Body)
			if err != nil {
				return
			}

			var n wp.WPNonce
			err = json.Unmarshal(nBytes, &n)
			if err != nil {
				return
			}

			url := fmt.Sprintf(
				"http://wordpress/wp-json/wp/v2/users/me?context=edit&_wpnonce=%v",
				n.Nonce,
			)
			uReq, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return
			}

			for _, cookie := range r.Cookies() {
				uReq.AddCookie(cookie)
			}

			uRes, err := client.Do(uReq)
			if err != nil || uRes.StatusCode != 200 {
				return
			}
			defer uRes.Body.Close()

			uBytes, err := io.ReadAll(uRes.Body)
			if err != nil {
				return
			}

			json.Unmarshal(uBytes, &u)
		}()

		// https://fideloper.com/golang-context-http-middleware
		// https://stackoverflow.com/a/70651651/7759523
		ctx := context.WithValue(r.Context(), "user", u)
		newReq := r.WithContext(ctx)
		next.ServeHTTP(w, newReq)
	})
}
