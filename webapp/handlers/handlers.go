package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"
	"webapp/embeds"
	"webapp/models"
	"webapp/wp"
	"webapp/wp/api"

	"github.com/gorilla/mux"
)

var homepageTmpl *template.Template
var postShowTmpl *template.Template
var tagShowTmpl *template.Template

var errorTmpl *template.Template

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
			ParseFS(
				embeds.Templates,
				"templates/_layout.tmpl",
				"templates/_header.tmpl",
				"templates/_footer.tmpl",
				"templates/_tag-list.tmpl",
				"templates/_tag-sidebar.tmpl",
				"templates/_post-list.tmpl",
			),
	)

	makeTmpl := func(file string) *template.Template {
		c := template.Must(tmplCommon.Clone())
		return template.Must(c.ParseFS(embeds.Templates, file))
	}

	homepageTmpl = makeTmpl("templates/post-index.tmpl")
	postShowTmpl = makeTmpl("templates/post-show.tmpl")
	tagShowTmpl = makeTmpl("templates/tag-show.tmpl")
	errorTmpl = makeTmpl(("templates/error.tmpl"))
}

func do404(w http.ResponseWriter, r *http.Request, msg string) {
	w.WriteHeader(404)

	title := "404 Error"
	err := errorTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   template.HTML(title),
		Request: *r,
		Data: map[string]any{
			"error": msg,
			"title": title,
			"user":  r.Context().Value("user"),
		},
	})
	if err != nil {
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
		return
	}
}

func do500(w http.ResponseWriter, r *http.Request, msg string) {
	w.WriteHeader(500)

	title := "500 Error"
	err := errorTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   template.HTML(title),
		Request: *r,
		Data: map[string]any{
			"error": msg,
			"title": title,
			"user":  r.Context().Value("user"),
		},
	})
	if err != nil {
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
		return
	}
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

	var errs []string
	if postErr != nil {
		errs = append(errs, fmt.Sprintf("Post error:\n%v", postErr.Error()))
	}
	if tagErr != nil {
		errs = append(errs, fmt.Sprintf("Tag error:\n%v", postErr.Error()))
	}
	if postErr != nil || tagErr != nil {
		do500(w, r, strings.Join(errs, "\n"))
		return
	}

	tagIdMap := map[int]wp.WPTag{}
	for _, tag := range *tags {
		tagIdMap[tag.Id] = tag
	}

	err := homepageTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   "Posts",
		Request: *r,
		Data: map[string]any{
			"posts":    posts,
			"tags":     tags,
			"tagIdMap": tagIdMap,
			"user":     r.Context().Value("user"),
		},
	})
	if err != nil {
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
		do500(w, r, fmt.Sprintf("Post error:\n %v", err.Error()))
		return
	}
	if len(*posts) == 0 {
		do404(w, r, "Post not found.\n")
		return
	}
	p := (*posts)[0]

	tags, _, err := api.Tags().
		SetParam("orderby", "name").
		SetParam("order", "asc").
		SetParam("include", p.Tags).
		GetAll()
	if err != nil {
		do500(w, r, fmt.Sprintf("There was an error.\n%v", err.Error()))
		return
	}
	tagIdMap := map[int]wp.WPTag{}
	for _, t := range *tags {
		tagIdMap[t.Id] = t
	}

	err = postShowTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   p.Title.Rendered,
		Request: *r,
		Data: map[string]any{
			"post":     p,
			"tagIdMap": tagIdMap,
			"user":     r.Context().Value("user"),
		},
	})
	if err != nil {
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
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
		do500(w, r, fmt.Sprintf("Tag error:\n%v", err.Error()))
		return
	}
	if len(*matchingTags) == 0 {
		do404(w, r, "Tag not found.\n")
		return
	}
	t := (*matchingTags)[0]

	posts, _, err := api.Posts().SetParam("tags", t.Id).GetAll()
	if err != nil {
		do500(w, r, fmt.Sprintf("There was an error.\n%v", err.Error()))
		return
	}

	var tags *[]wp.WPTag
	tags, _, err = api.Tags().
		SetParam("orderby", "name").
		SetParam("order", "asc").
		GetAll()
	if err != nil {
		do500(w, r, fmt.Sprintf("There was an error.\n%v", err.Error()))
		return
	}
	tagIdMap := map[int]wp.WPTag{}
	for _, t := range *tags {
		tagIdMap[t.Id] = t
	}

	err = tagShowTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
		Title:   template.HTML(fmt.Sprintf("Tag: %v", t.Name)),
		Request: *r,
		Data: map[string]any{
			"tag":      t,
			"tagIdMap": tagIdMap,
			"tags":     tags,
			"tagSlug":  slug,
			"posts":    posts,
			"user":     r.Context().Value("user"),
		},
	})
	if err != nil {
		fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	do404(w, r, "Page not found.\n")
}
