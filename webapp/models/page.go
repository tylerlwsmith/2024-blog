package models

import (
	"html/template"
	"net/http"
)

type PageData struct {
	Request http.Request
	Title   template.HTML
	Data    any
}
