package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grigoryan-vl/url-shortener/internal/config"
	handler "github.com/grigoryan-vl/url-shortener/internal/handler/short-url"
	urlService "github.com/grigoryan-vl/url-shortener/internal/service/short-url"
)

func main() {
	cfg := config.ParseFlags()

	urlSvc := urlService.NewUrlService()
	handler := handler.NewHandler(urlSvc, strings.TrimSpace(cfg.BaseUrl))

	r := chi.NewRouter()
	r.HandleFunc("/", handler.CreateShortUrlHandler)
	r.HandleFunc("/{id}", handler.GetUrlByIdHandler)

	err := http.ListenAndServe(cfg.Addr, r)
	if err != nil {
		panic(err)
	}
}
