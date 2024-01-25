package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"webapp/models"
	"webapp/wp"
	"webapp/wp/api"
)

var homepageTmpl *template.Template
var postsTmpl *template.Template
var tagTmpl *template.Template

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
				"templates/_post-list.tmpl",
			),
	)

	makeTmpl := func(file string) *template.Template {
		c := template.Must(tmplCommon.Clone())
		return template.Must(c.ParseFiles(file))
	}

	homepageTmpl = makeTmpl("templates/post-index.tmpl")
	postsTmpl = makeTmpl("templates/post-show.tmpl")
	tagTmpl = makeTmpl("templates/tag-show.tmpl")
}

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
				fmt.Println(cookie)
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

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Use(StripTrailingSlashesMiddleware)
	r.Use(UserMiddleware)

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
			Data:    map[string]any{"posts": posts, "tags": tags, "tagIdMap": tagIdMap, "user": r.Context().Value("user")},
		})

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
		}
	})

	r.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.HandleFunc("/posts/{slug}", func(w http.ResponseWriter, r *http.Request) {
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

		tags, _, err := api.Tags().SetParam("include", p.Tags).GetAll()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "There was an error.\n", err.Error())
			return
		}
		tagIdMap := map[int]wp.WPTag{}
		for _, t := range *tags {
			tagIdMap[t.Id] = t
		}

		err = postsTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   p.Title.Rendered,
			Request: *r,
			Data:    map[string]any{"post": p, "tagIdMap": tagIdMap, "user": r.Context().Value("user")},
		})

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
			return
		}
	})

	r.HandleFunc("/tags/{slug}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := vars["slug"]

		tags, _, err := api.Tags().
			SetParam("slug", slug).
			SetParam("per_page", 1).
			Get()

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "tag error:\n %v", err.Error())
			return
		}
		if len(*tags) == 0 {
			w.WriteHeader(404)
			fmt.Fprint(w, "Post not found.\n")
			fmt.Fprint(w, "http://wordpress:80/wp-json/wp/v2/posts?slug="+slug)
			return
		}
		t := (*tags)[0]

		posts, _, err := api.Posts().SetParam("tags", t.Id).GetAll()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "there was an error.\n", err.Error())
			return
		}

		postTagIds := []int{}
		for _, p := range *posts {
			postTagIds = append(postTagIds, p.Tags...)
		}
		// Dedupe IDs. https://stackoverflow.com/a/76471309/7759523
		slices.Sort(postTagIds) // mutates original slice.
		uniqTagIds := slices.Compact(postTagIds)

		postTags, _, err := api.Tags().SetParam("include", uniqTagIds).GetAll()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "there was an error.\n", err.Error())
			return
		}
		tagIdMap := map[int]wp.WPTag{}
		for _, t := range *postTags {
			tagIdMap[t.Id] = t
		}

		err = tagTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   template.HTML(fmt.Sprintf("Tag: %v", t.Name)),
			Request: *r,
			Data:    map[string]any{"tag": t, "posts": posts, "tagIdMap": tagIdMap, "user": r.Context().Value("user")},
		})

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "There was an error executing the templates.\n", err.Error())
			return
		}
	})

	// Middleware is typically skipped when there is no matching route. Our app
	// will strip trailing slashes so we need a custom NotFoundHandler.
	// https://github.com/gorilla/mux/issues/636
	// https://stackoverflow.com/a/56937571/7759523
	r.NotFoundHandler = r.NewRoute().HandlerFunc(http.NotFound).GetHandler()
	http.ListenAndServe(":3000", r)
}
