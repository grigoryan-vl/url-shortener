package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grigoryan-vl/URL-shortener/internal/config"
	pingHndl "github.com/grigoryan-vl/URL-shortener/internal/handler/ping"
	urlHndl "github.com/grigoryan-vl/URL-shortener/internal/handler/short-url"
	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
	"github.com/grigoryan-vl/URL-shortener/internal/repository/file"
	"github.com/grigoryan-vl/URL-shortener/internal/repository/memory"
	"github.com/grigoryan-vl/URL-shortener/internal/repository/postgres"
	pingService "github.com/grigoryan-vl/URL-shortener/internal/service/ping"
	URLService "github.com/grigoryan-vl/URL-shortener/internal/service/short-url"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	cfg := config.ParseFlags()
	ctx := context.Background()

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	// 1. Один пул — только если задан DSN явно (env или флаг -d)
	var pool *pgxpool.Pool
	if cfg.DBConfig.DatabaseDSN != "" {
		pool, err = pgxpool.New(ctx, cfg.DBConfig.DatabaseDSN)
		if err != nil {
			sugar.Fatalw("failed to create pool", "error", err)
		}
		defer pool.Close()

		if err := pool.Ping(ctx); err != nil {
			sugar.Warnw("db unavailable at start, /ping will return 500", "error", err)
		}
	}

	// 2. Storage по приоритету: postgres → file → memory
	var st storage.Storage
	switch {
	case pool != nil:
		st, err = postgres.New(ctx, pool)
	case cfg.FileStoragePath != "":
		st, err = file.New(cfg.FileStoragePath)
	default:
		st = memory.New()
	}
	if err != nil {
		sugar.Fatalw("failed to init storage", "error", err)
	}
	defer st.Close()

	// 3. Сервисы и хендлеры
	urlSvc := URLService.NewURLService(st)
	urlHandler := urlHndl.NewURLHandler(urlSvc, strings.TrimSpace(cfg.BaseURL))

	pingSvc := pingService.NewPingService(pool)
	pingHandler := pingHndl.NewPingHandler(pingSvc)

	r := chi.NewRouter()
	r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(urlHandler.CreateShortURLHandler()), sugar).ServeHTTP)
	r.Post("/api/shorten", middleware.WithLogging(middleware.GzipMiddleware(urlHandler.CreateShortURLHandlerV2()), sugar).ServeHTTP)
	r.Get("/{id}", middleware.WithLogging(urlHandler.GetURLByIDHandler(), sugar).ServeHTTP)
	r.Get("/ping", middleware.WithLogging(pingHandler.Ping(), sugar).ServeHTTP)

	sugar.Infow("Starting server", "addr", cfg.Addr)

	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
