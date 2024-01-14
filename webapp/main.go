package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"webapp/models"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		postResChan := make(chan *http.Response)
		postErrChan := make(chan error)
		go func(rc chan *http.Response, ec chan error) {
			resp, err := http.Get("http://wordpress:80/wp-json/wp/v2/posts")
			rc <- resp
			ec <- err
		}(postResChan, postErrChan)

		tagResChan := make(chan *http.Response)
		tagErrChan := make(chan error)
		go func(rc chan *http.Response, ec chan error) {
			resp, err := http.Get("http://wordpress:80/wp-json/wp/v2/tags")
			rc <- resp
			ec <- err
		}(tagResChan, tagErrChan)

		postResp, postErr := <-postResChan, <-postErrChan
		// tagResp, tagErr := <-postResChan, <-postErrChan

		if postErr != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "WordPress failed to return a response.\n", postErr.Error())
			return
		}

		if postResp.StatusCode > 299 {
			w.WriteHeader(503)
			fmt.Fprint(w, "WordPress returned a non-200 status code.")
			return
		}

		bodyStr, err := io.ReadAll(postResp.Body)
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error in reading the body of the WordPress response.\n")
			fmt.Fprint(w, err.Error())
			return
		}

		err = postResp.Body.Close()
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "The body response could not be closed.\n", err.Error())
			return
		}

		posts := []models.WPPost{}
		err = json.Unmarshal(bodyStr, &posts)
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error unmarshalling JSON.\n", err.Error())
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/post-index.tmpl",
			"templates/_layout.tmpl",
			"templates/_header.tmpl",
			"templates/_footer.tmpl",
		)

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error parsing the templates.\n", err.Error())
			return
		}

		err = tmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   "Posts",
			Request: *r,
			Data:    posts,
		})

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
			return
		}
	})

	r.HandleFunc("/posts/{slug}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := vars["slug"]

		resp, err := http.Get("http://wordpress:80/wp-json/wp/v2/posts?slug=" + slug)
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "WordPress failed to return a response.\n", err.Error())
			return
		}

		if resp.StatusCode > 299 {
			w.WriteHeader(503)
			fmt.Fprint(w, "WordPress returned a non-200 status code.\n")
		}

		bodyStr, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error in reading the body of the WordPress response.\n")
			fmt.Fprint(w, err.Error())
			return
		}

		err = resp.Body.Close()
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "The body response could not be closed.\n", err.Error())
			return
		}

		posts := []models.WPPost{}
		err = json.Unmarshal(bodyStr, &posts)

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error unmarshalling JSON.\n", err.Error())
			return
		}

		if len(posts) == 0 {
			w.WriteHeader(404)
			fmt.Fprint(w, "Post not found.\n")
			fmt.Fprint(w, "http://wordpress:80/wp-json/wp/v2/posts?slug="+slug)
			return

		}
		p := posts[0]

		tmpl, err := template.ParseFiles(
			"templates/post-show.tmpl",
			"templates/_layout.tmpl",
			"templates/_header.tmpl",
			"templates/_footer.tmpl",
		)

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error parsing the templates.\n", err.Error())
			return
		}

		err = tmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   p.Title.Rendered,
			Request: *r,
			Data:    p,
		})

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
			return
		}
	})

	http.ListenAndServe(":3000", r)
}
