package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	handler "github.com/grigoryan-vl/url-shortener/internal/handler/short-url"
	urlService "github.com/grigoryan-vl/url-shortener/internal/service/short-url"
)

func main() {
	//init
	urlService := urlService.NewUrlService()
	handler := handler.NewHandler(urlService)

	r := chi.NewRouter()
	r.HandleFunc("/", handler.CreateShortUrlHandler)
	r.HandleFunc("/{id}", handler.GetUrlByIdHandler)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
