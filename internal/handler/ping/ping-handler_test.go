package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
)

// stubPingService — заглушка для PingService.
type stubPingService struct {
	err error
}

func (s *stubPingService) Ping(_ context.Context) error {
	return s.err
}

func TestPingHandler(t *testing.T) {
	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar := *logger.Sugar()

	testCases := []struct {
		name          string
		method        string
		path          string
		svcErr        error
		expectedCode  int
		expectedAllow string
	}{
		{
			name:          "method not allowed",
			method:        http.MethodPost,
			path:          "/ping",
			expectedCode:  http.StatusMethodNotAllowed,
			expectedAllow: http.MethodGet,
		},
		{
			name:         "success",
			method:       http.MethodGet,
			path:         "/ping",
			expectedCode: http.StatusOK,
		},
		{
			name:         "service error",
			method:       http.MethodGet,
			path:         "/ping",
			svcErr:       errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubPingService{err: tc.svcErr}
			handler := NewPingHandler(svc)

			r := chi.NewRouter()
			r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			})
			r.Get("/ping", middleware.WithLogging(handler.Ping(), sugar).ServeHTTP)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code, "код ответа не совпадает")
			if tc.expectedAllow != "" {
				assert.Equal(t, tc.expectedAllow, rr.Header().Get("Allow"))
			}
		})
	}
}
