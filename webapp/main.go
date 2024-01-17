package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"text/template"

	"github.com/gorilla/mux"

	"webapp/models"
	"webapp/wp"
	"webapp/wp/api"
)

var tmplCommon = []string{"templates/_layout.tmpl", "templates/_header.tmpl", "templates/_footer.tmpl"}
var homepageTmpl = template.Must(template.ParseFiles(append(tmplCommon, "templates/post-index.tmpl")...))
var postsTmpl = template.Must(template.ParseFiles(append(tmplCommon, "templates/post-show.tmpl")...))

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var posts []wp.WPPost
		var tags []wp.WPTag
		var postErr, tagErr error

		wg := sync.WaitGroup{}

		wg.Add(2)
		go func() { defer wg.Done(); posts, _, postErr = api.Posts().Get() }()
		go func() { defer wg.Done(); tags, _, tagErr = api.Tags().Get() }()
		wg.Wait()

		if postErr != nil {
			fmt.Fprintf(w, "post error:\n %v", postErr.Error())
		}
		if tagErr != nil {
			fmt.Fprintf(w, "tag error:\n %v", tagErr.Error())
		}
		if postErr != nil || tagErr != nil {
			w.WriteHeader(503)
			return
		}

		err := homepageTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   "Posts",
			Request: *r,
			Data:    map[string]any{"posts": posts, "tags": tags},
		})

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
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

		posts := []wp.WPPost{}
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

		err = postsTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
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
