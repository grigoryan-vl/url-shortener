package middleware

import (
	"net/http"
	"time"

	models "github.com/grigoryan-vl/url-shortener/internal/model/middleware"
	"go.uber.org/zap"
)

func WithLogging(h http.Handler, sugar zap.SugaredLogger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		responseData := &models.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := models.LoggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			ResponseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		sugar.Infoln(
			"uri", uri,
			"method", method,
			"status", responseData.Status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.Size, // получаем перехваченный размер ответа
		)
	}

	return http.HandlerFunc(logFn)
}
