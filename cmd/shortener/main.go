package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grigoryan-vl/URL-shortener/internal/config"
	handler "github.com/grigoryan-vl/URL-shortener/internal/handler/short-url"
	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
	URLService "github.com/grigoryan-vl/URL-shortener/internal/service/short-url"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	cfg := config.ParseFlags()

	URLSvc, err := URLService.NewURLService(cfg.FileStoragePath)
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	handler := handler.NewHandler(URLSvc, strings.TrimSpace(cfg.BaseURL))

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
	r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(handler.CreateShortURLHandler()), sugar).ServeHTTP)
	r.Post("/api/shorten", middleware.WithLogging(middleware.GzipMiddleware(handler.CreateShortURLHandlerV2()), sugar).ServeHTTP)
	r.Get("/{id}", middleware.WithLogging(handler.GetURLByIDHandler(), sugar).ServeHTTP)

	sugar.Infow(
		"Starting server",
		"addr", cfg.Addr,
	)

	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
