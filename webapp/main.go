package main

import (
	"fmt"
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
		var posts *[]wp.WPPost
		var tags *[]wp.WPTag
		var postErr, tagErr error

		wg := sync.WaitGroup{}

		wg.Add(2)
		go func() { defer wg.Done(); posts, _, postErr = api.Posts().GetAll() }()
		go func() { defer wg.Done(); tags, _, tagErr = api.Tags().GetAll() }()
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

		tagIdMap := map[int]wp.WPTag{}
		for _, tag := range *tags {
			tagIdMap[tag.Id] = tag
		}

		err := homepageTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   "Posts",
			Request: *r,
			Data:    map[string]any{"posts": posts, "tags": tags, "tagIdMap": tagIdMap},
		})

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
		}
	})

	r.HandleFunc("/posts/{slug}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := vars["slug"]

		posts, _, err := api.Posts().
			SetParam("slug", slug).
			SetParam("per_page", 1).
			Get()

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprintf(w, "post error:\n %v", err.Error())
			return
		}
		if len(*posts) == 0 {
			w.WriteHeader(404)
			fmt.Fprint(w, "Post not found.\n")
			fmt.Fprint(w, "http://wordpress:80/wp-json/wp/v2/posts?slug="+slug)
			return
		}
		p := (*posts)[0]

		t, _, err := api.Tags().SetParam("include", p.Tags).GetAll()
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
			return
		}
		tagIdMap := map[int]wp.WPTag{}
		for _, tag := range *t {
			tagIdMap[tag.Id] = tag
		}

		err = postsTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   p.Title.Rendered,
			Request: *r,
			Data:    map[string]any{"post": p, "tagIdMap": tagIdMap},
		})

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
			return
		}
	})

	http.ListenAndServe(":3000", r)
}
