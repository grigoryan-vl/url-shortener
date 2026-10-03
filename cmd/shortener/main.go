package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grigoryan-vl/URL-shortener/internal/config"
	pingHndl "github.com/grigoryan-vl/URL-shortener/internal/handler/ping"
	urlHndl "github.com/grigoryan-vl/URL-shortener/internal/handler/short-url"
	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
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

	URLSvc, err := URLService.NewURLService(cfg.FileStoragePath)
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	urlHandler := urlHndl.NewURLHandler(URLSvc, strings.TrimSpace(cfg.BaseURL))

	dsn := cfg.DBConfig.DBDSN
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
			cfg.DBConfig.Login,
			cfg.DBConfig.Password,
			cfg.DBConfig.Host,
			cfg.DBConfig.Port,
			cfg.DBConfig.DBName)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		sugar.Fatalw("не удалось создать пул соединений", "error", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		sugar.Warnw("БД недоступна при старте, /ping будет возвращать 500", "error", err)
	}

	svc := pingService.NewPingService(pool)
	pingHandler := pingHndl.NewPingHandler(svc)

	r := chi.NewRouter()
	r.Post("/", middleware.WithLogging(middleware.GzipMiddleware(urlHandler.CreateShortURLHandler()), sugar).ServeHTTP)
	r.Post("/api/shorten", middleware.WithLogging(middleware.GzipMiddleware(urlHandler.CreateShortURLHandlerV2()), sugar).ServeHTTP)
	r.Get("/{id}", middleware.WithLogging(urlHandler.GetURLByIDHandler(), sugar).ServeHTTP)
	r.Get("/ping", middleware.WithLogging(pingHandler.Ping(), sugar).ServeHTTP)

	sugar.Infow(
		"Starting server",
		"addr", cfg.Addr,
	)

	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}
