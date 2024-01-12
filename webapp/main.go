package main

import (
	"encoding/json"
	"fmt"
	_ "html/template"
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
			fmt.Fprint(w, "Can't connect to WordPress")
			return
		}

		bodyStr, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			w.WriteHeader(503)
			fmt.Fprint(w, "WordPress returned a non-200 status code.")
			return
		}

		if err != nil {
			w.WriteHeader(503)
			fmt.Fprint(w, "There was an error in reading the body of the WordPress response")
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
			fmt.Fprint(w, "There was an error unmarshalling JSON.\n")
			fmt.Fprint(w, err.Error())
			return
		}

		fmt.Fprintln(w, posts[0].Title.Rendered)
		fmt.Fprintln(w, posts[0].Content.Rendered)
		fmt.Fprintln(w, posts[0].Date)
		fmt.Fprintln(w, posts[0].Link.String())
	})

	http.ListenAndServe(":3000", r)
}
