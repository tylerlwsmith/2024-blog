package main

import (
	"encoding/json"
	"fmt"
	_ "html/template"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// type WPPage struct {
// 	Id int `json:"id"`
// }

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

		var bodyData []map[string]interface{}
		json.Unmarshal(bodyStr, &bodyData)

		// https://go.dev/tour/methods/15
		c := bodyData[0]["content"].(map[string]interface{})

		// jsonContent, err := json.Marshal(c["rendered"])
		// if err != nil {
		// 	fmt.Fprint(w, "Error Marshalling JSON")
		// 	return
		// }

		fmt.Fprint(w, c["rendered"])
	})

	http.ListenAndServe(":3000", r)
}
