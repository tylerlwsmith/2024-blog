package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/gorilla/mux"

	"webapp/helpers"
	"webapp/models"
)

var tmplCommon = []string{"templates/_layout.tmpl", "templates/_header.tmpl", "templates/_footer.tmpl"}
var homepageTmpl = template.Must(template.ParseFiles(append(tmplCommon, "templates/post-index.tmpl")...))
var postsTmpl = template.Must(template.ParseFiles(append(tmplCommon, "templates/post-show.tmpl")...))

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var postResp, tagResp *http.Response
		var postErr, tagErr error

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			postResp, postErr = http.Get("http://wordpress:80/wp-json/wp/v2/posts")
		}()

		go func() {
			defer wg.Done()
			tagResp, tagErr = http.Get("http://wordpress:80/wp-json/wp/v2/tags")
		}()

		wg.Wait()

		// Order is important: all error checking below follows this order.
		resTypes := [2]string{"post", "tag"}
		responses := [2]*http.Response{postResp, tagResp}
		resErrs := [2]error{postErr, tagErr}

		resErrMsgs := []string{}

		// Body **must** be closed, but only on respones with no error.
		// https://stackoverflow.com/a/32819910/7759523
		for i, err := range resErrs {
			if err == nil {
				defer responses[i].Body.Close()
			} else {
				resErrMsgs = append(resErrMsgs, err.Error())
			}
		}

		if len(resErrMsgs) > 0 {
			fmt.Fprint(w, strings.Join(resErrMsgs[:], "\n"))
			return
		}

		for i, res := range responses {
			if res.StatusCode > 299 {
				resErrMsgs = append(
					resErrMsgs,
					fmt.Sprintf("The %v endpoint returned a non-200 status code.\n", resTypes[i]),
				)
			}
		}

		if len(resErrMsgs) > 0 {
			fmt.Fprint(w, strings.Join(resErrMsgs[:], "\n"))
			return
		}

		posts := []models.WPPost{}
		tags := []models.WPTag{}
		postErr = helpers.UnmarshalJsonReader[[]models.WPPost](postResp.Body, &posts)
		tagErr = helpers.UnmarshalJsonReader[[]models.WPTag](tagResp.Body, &tags)

		for i, err := range []error{postErr, tagErr} {
			if err != nil {
				resErrMsgs = append(
					resErrMsgs,
					fmt.Sprintf("%v error: %v", resTypes[i], err.Error()),
				)
			}
		}

		if len(resErrMsgs) > 0 {
			fmt.Fprint(w, strings.Join(resErrMsgs[:], "\n"))
			return
		}

		err := homepageTmpl.ExecuteTemplate(w, "_layout.tmpl", models.PageData{
			Title:   "Posts",
			Request: *r,
			Data: map[string]any{
				"posts": posts,
				"tags":  tags,
			},
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
