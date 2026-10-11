package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
	"github.com/grigoryan-vl/URL-shortener/internal/repository/file"
	URLServ "github.com/grigoryan-vl/URL-shortener/internal/service/short-url"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestService(t *testing.T) *URLServ.URLService {
	t.Helper()
	path := filepath.Join(t.TempDir(), "storage.json")
	st, err := file.New(path)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}
	return URLServ.NewURLService(st)
}

func TestCreateShortURL(t *testing.T) {
	handler := NewURLHandler(newTestService(t), "")

	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	r.Post("/", handler.CreateShortURLHandler().ServeHTTP)

	testCases := []struct {
		name          string
		method        string
		contentType   string
		body          string
		expectedCode  int
		expectedAllow string
		bodyContains  string
	}{
		{
			name:          "method not allowed",
			method:        http.MethodGet,
			contentType:   "text/plain",
			body:          "https://example.com",
			expectedCode:  http.StatusMethodNotAllowed,
			expectedAllow: http.MethodPost,
		},
		{
			name:         "wrong content type",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{"URL":"https://example.com"}`,
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:         "empty body",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "whitespace body",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "   \n\t  ",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "success",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "https://example.com",
			expectedCode: http.StatusCreated,
			bodyContains: "http://",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code, "код ответа не совпадает")
			if tc.expectedAllow != "" {
				assert.Equal(t, tc.expectedAllow, rr.Header().Get("Allow"))
			}
			if tc.bodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.bodyContains, "тело ответа не содержит ожидаемую строку")
			}
		})
	}
}

func TestCreateShortURLV2(t *testing.T) {
	handler := NewURLHandler(newTestService(t), "")

	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	r.Post("/", handler.CreateShortURLHandlerV2().ServeHTTP)

	testCases := []struct {
		name          string
		method        string
		contentType   string
		body          string
		expectedCode  int
		expectedAllow string
		bodyContains  string
	}{
		{
			name:         "success",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{"url":"https://example.com"}`,
			expectedCode: http.StatusCreated,
			bodyContains: `"result"`,
		},
		{
			name:          "method not allowed",
			method:        http.MethodGet,
			contentType:   "application/json",
			body:          `{"url":"https://example.com"}`,
			expectedCode:  http.StatusMethodNotAllowed,
			expectedAllow: http.MethodPost,
			bodyContains:  "Method Not Allowed",
		},
		{
			name:         "wrong media type",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         `{"url":"https://example.com"}`,
			expectedCode: http.StatusUnsupportedMediaType,
			bodyContains: "application/json",
		},
		{
			name:         "empty body",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         "",
			expectedCode: http.StatusBadRequest,
			bodyContains: "Failed to unmarshal body",
		},
		{
			name:         "whitespace body",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         "   \n\t  ",
			expectedCode: http.StatusBadRequest,
			bodyContains: "Failed to unmarshal body",
		},
		{
			name:         "malformed json",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{"url":`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "Failed to unmarshal body",
		},
		{
			name:         "missing URL field",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{}`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "URL is required",
		},
		{
			name:         "empty URL",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{"url":""}`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "URL is required",
		},
		{
			name:         "whitespace URL",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{"url":"   \n\t  "}`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "URL is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code, "код ответа не совпадает")
			if tc.expectedAllow != "" {
				assert.Equal(t, tc.expectedAllow, rr.Header().Get("Allow"))
			}
			if tc.bodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.bodyContains, "тело ответа не содержит ожидаемую строку")
			}
		})
	}
}

var sugar zap.SugaredLogger

func TestGetURLByIDHandler(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	existingCode, err := svc.CreateShortURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("failed to create short url: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	r.Get("/{id}", middleware.WithLogging(handler.GetURLByIDHandler(), sugar).ServeHTTP)

	testCases := []struct {
		name          string
		method        string
		path          string
		expectedCode  int
		expectedAllow string
		expectedLoc   string
	}{
		{
			name:          "method not allowed",
			method:        http.MethodPost,
			path:          "/abc",
			expectedCode:  http.StatusMethodNotAllowed,
			expectedAllow: http.MethodGet,
		},
		{
			name:         "empty id",
			method:       http.MethodGet,
			path:         "/",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "not found",
			method:       http.MethodGet,
			path:         "/nonexistent",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "success redirect",
			method:       http.MethodGet,
			path:         "/" + existingCode,
			expectedCode: http.StatusTemporaryRedirect,
			expectedLoc:  "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code, "код ответа не совпадает")
			if tc.expectedAllow != "" {
				assert.Equal(t, tc.expectedAllow, rr.Header().Get("Allow"))
			}
			if tc.expectedLoc != "" {
				assert.Equal(t, tc.expectedLoc, rr.Header().Get("Location"))
			}
		})
	}
}
