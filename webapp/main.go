package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"webapp/embeds"
	"webapp/handlers"
	"webapp/middleware"
)

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(embeds.Static)))

	r.Use(middleware.StripTrailingSlashesMiddleware)
	r.Use(middleware.UserMiddleware)

	r.HandleFunc("/", handlers.Homepage)

	r.HandleFunc("/posts", handlers.PostIndex)
	r.HandleFunc("/posts/{slug}", handlers.PostShow)

	r.HandleFunc("/tags", handlers.TagIndex)
	r.HandleFunc("/tags/{slug}", handlers.TagShow)

	// Middleware is typically skipped when there is no matching route. Our app
	// will strip trailing slashes so we need a custom NotFoundHandler.
	// https://github.com/gorilla/mux/issues/636
	// https://stackoverflow.com/a/56937571/7759523
	r.NotFoundHandler = r.NewRoute().HandlerFunc(handlers.NotFoundHandler).GetHandler()

	fmt.Println("Starting webapp server...")
	http.ListenAndServe(":3000", r)
}
