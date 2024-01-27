package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"
	"webapp/models"
	"webapp/wp"
	"webapp/wp/api"

	"github.com/gorilla/mux"
)

var homepageTmpl *template.Template
var postShowTmpl *template.Template
var tagShowTmpl *template.Template

// var http404Tmpl *template.Template
// var http500Tmpl *template.Template

func init() {
	var tmplCommon = template.Must(
		template.
			New("common").
			Funcs(template.FuncMap{
				"currentYear": func() string {
					return time.Now().Format("2006")
				},
				"args": func(args ...any) []any {
					return args
				},
			}).
			ParseFiles(
				"templates/_layout.tmpl",
				"templates/_header.tmpl",
				"templates/_footer.tmpl",
				"templates/_sidebar.tmpl",
				"templates/_post-list.tmpl",
			),
	)

	makeTmpl := func(file string) *template.Template {
		c := template.Must(tmplCommon.Clone())
		return template.Must(c.ParseFiles(file))
	}

	homepageTmpl = makeTmpl("templates/post-index.tmpl")
	postShowTmpl = makeTmpl("templates/post-show.tmpl")
	tagShowTmpl = makeTmpl("templates/tag-show.tmpl")
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	var posts *[]wp.WPPost
	var tags *[]wp.WPTag
	var postErr, tagErr error

	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		posts, _, postErr = api.Posts().
			SetParam("orderby", "date").
			SetParam("order", "desc").
			GetAll()
	}()
	go func() {
		defer wg.Done()
		tags, _, tagErr = api.Tags().
			SetParam("orderby", "name").
			SetParam("order", "asc").
			GetAll()
	}()
	wg.Wait()

	if postErr != nil {
		fmt.Fprintf(w, "post error:\n %v", postErr.Error())
	}
	if tagErr != nil {
		fmt.Fprintf(w, "tag error:\n %v", tagErr.Error())
	}
	if postErr != nil || tagErr != nil {
		w.WriteHeader(500)
		return
	}

	tagIdMap := map[int]wp.WPTag{}
	for _, tag := range *tags {
		tagIdMap[tag.Id] = tag
	}

	err := homepageTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   "Posts",
		Request: *r,
		Data:    map[string]any{
			"posts": posts,
			"tags": tags,
			"tagIdMap": tagIdMap,
			"user": r.Context().Value("user")
		},
	})

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
	}
}

func PostIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func PostShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	posts, _, err := api.Posts().
		SetParam("slug", slug).
		SetParam("per_page", 1).
		Get()

	if err != nil {
		w.WriteHeader(500)
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

	tags, _, err := api.Tags().
		SetParam("orderby", "name").
		SetParam("order", "asc").
		SetParam("include", p.Tags).
		GetAll()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "There was an error.\n", err.Error())
		return
	}
	tagIdMap := map[int]wp.WPTag{}
	for _, t := range *tags {
		tagIdMap[t.Id] = t
	}

	err = postShowTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   p.Title.Rendered,
		Request: *r,
		Data:    map[string]any{
			"post": p,
			"tagIdMap": tagIdMap,
			"user": r.Context().Value("user")
		},
	})

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
		return
	}
}

func TagIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func TagShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	matchingTags, _, err := api.Tags().
		SetParam("slug", slug).
		SetParam("per_page", 1).
		Get()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "tag error:\n %v", err.Error())
		return
	}
	if len(*matchingTags) == 0 {
		w.WriteHeader(404)
		fmt.Fprint(w, "Post not found.\n")
		fmt.Fprint(w, "http://wordpress:80/wp-json/wp/v2/posts?slug="+slug)
		return
	}
	t := (*matchingTags)[0]

	posts, _, err := api.Posts().SetParam("tags", t.Id).GetAll()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "there was an error.\n", err.Error())
		return
	}

	var tags *[]wp.WPTag
	tags, _, err = api.Tags().
		SetParam("orderby", "name").
		SetParam("order", "asc").
		GetAll()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "there was an error.\n", err.Error())
		return
	}
	tagIdMap := map[int]wp.WPTag{}
	for _, t := range *tags {
		tagIdMap[t.Id] = t
	}

	err = tagShowTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   template.HTML(fmt.Sprintf("Tag: %v", t.Name)),
		Request: *r,
		Data:    map[string]any{
			"tag": t,
			"tags": tags,
			"tagIdMap": tagIdMap,
			"posts": posts,
			"user": r.Context().Value("user")
		},
	})

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
		return
	}
}
