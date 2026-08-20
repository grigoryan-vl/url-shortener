package main

import (
	"net/http"

	handler "github.com/grigoryan-vl/url-shortener/internal/handler/short-url"
	urlService "github.com/grigoryan-vl/url-shortener/internal/service/short-url"
)

func main() {
	//init
	urlService := urlService.NewUrlService()
	handler := handler.NewHandler(urlService)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.CreateShortUrlHandler)
	mux.HandleFunc(`/{id}`, handler.GetUrlByIdHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
