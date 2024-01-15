package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"

	"webapp/async"
	"webapp/models"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		postReq := async.Get("http://wordpress:80/wp-json/wp/v2/posts")
		tagReq := async.Get("http://wordpress:80/wp-json/wp/v2/tags")

		postResp, postErr := postReq.AwaitResponse()
		tagResp, tagErr := tagReq.AwaitResponse()

		// Order is important: all error checking below follows this order.
		resTypes := [2]string{"post", "tag"}
		responses := [2]*http.Response{postResp, tagResp}
		resErrs := [2]error{postErr, tagErr}

		resErrMsgs := []string{}

		// Body **must** only be closed on respones with no error.
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

		bodyStr, err := io.ReadAll(postResp.Body)
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error in reading the body of the WordPress response.\n")
			fmt.Fprint(w, err.Error())
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
