package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grigoryan-vl/url-shortener/internal/config"
	handler "github.com/grigoryan-vl/url-shortener/internal/handler/short-url"
	middleware "github.com/grigoryan-vl/url-shortener/internal/middleware"
	urlService "github.com/grigoryan-vl/url-shortener/internal/service/short-url"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	cfg := config.ParseFlags()

	urlSvc := urlService.NewUrlService()
	handler := handler.NewHandler(urlSvc, strings.TrimSpace(cfg.BaseUrl))

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	r := chi.NewRouter()
	r.Handle("/", middleware.WithLogging(handler.CreateShortUrlHandler(), sugar))
	r.Handle("/api/shorten", middleware.WithLogging(handler.CreateShortUrlHandlerV2(), sugar))
	r.Handle("/{id}", middleware.WithLogging(handler.GetUrlByIdHandler(), sugar))

	sugar.Infow(
		"Starting server",
		"addr", cfg.Addr,
	)

	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
