package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"webapp/wp"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get("http://wordpress:80/wp-json/wp/v2/posts")
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "Can't connect to WordPress", err.Error())
			return
		}

		if res.StatusCode > 299 {
			w.WriteHeader(503)
			fmt.Fprint(w, "WordPress returned a non-200 status code.")
			return
		}

		bodyStr, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error in reading the body of the WordPress response\n")
			fmt.Fprint(w, err.Error())
			return
		}

		////////// START ORIGINAL IMPLEMENTATION

		// var bodyData []map[string]interface{}
		// json.Unmarshal(bodyStr, &bodyData)

		// // https://go.dev/tour/methods/15
		// c := bodyData[0]["content"].(map[string]interface{})

		// // jsonContent, err := json.Marshal(c["rendered"])
		// // if err != nil {
		// // 	fmt.Fprint(w, "Error Marshalling JSON")
		// // 	return
		// // }

		// fmt.Fprint(w, c["rendered"])

		////////// END ORIGINAL IMPLEMENTATION

		posts := []wp.WPPost{}
		err = json.Unmarshal(bodyStr, &posts)
		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error unmarshalling JSON.\n", err.Error())
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/post-index.tmpl",
			"templates/_header.tmpl",
			"templates/_footer.tmpl",
		)

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error loading the template.\n", err.Error())
			return
		}

		tmpl.ExecuteTemplate(w, "post-index.tmpl", posts)
	})

	http.ListenAndServe(":3000", r)
}
