package main

import (
	"net/http"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/handler"
)

func main() {
	
	urlService := service.NewURLService()

	mux := http.NewServeMux()
	mux.Handle("/", handler.PostHandler(urlService))
	mux.Handle("/{id}", handler.GetHandler(urlService))
	
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
