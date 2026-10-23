package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
	model "github.com/grigoryan-vl/URL-shortener/internal/model/request"
	"github.com/grigoryan-vl/URL-shortener/internal/repository/file"
	URLServ "github.com/grigoryan-vl/URL-shortener/internal/service/short-url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCreateShortURLBatch(t *testing.T) {
	handler := NewURLHandler(newTestService(t), "")

	r := chi.NewRouter()
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	r.Post("/api/shorten/batch", handler.CreateShortURLBatchHandler().ServeHTTP)

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
			contentType:   "application/json",
			body:          `[{"correlation_id":"1","original_url":"https://example.com"}]`,
			expectedCode:  http.StatusMethodNotAllowed,
			expectedAllow: http.MethodPost,
		},
		{
			name:         "wrong media type",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         `[{"correlation_id":"1","original_url":"https://example.com"}]`,
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
			name:         "malformed json",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `[{"correlation_id":`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "Failed to unmarshal body",
		},
		{
			name:         "empty batch",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `[]`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "Empty batch",
		},
		{
			name:         "empty original_url",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `[{"correlation_id":"1","original_url":""}]`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "URL is required",
		},
		{
			name:         "whitespace original_url",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `[{"correlation_id":"1","original_url":"   "}]`,
			expectedCode: http.StatusBadRequest,
			bodyContains: "URL is required",
		},
		{
			name:         "success single",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `[{"correlation_id":"1","original_url":"https://example.com"}]`,
			expectedCode: http.StatusCreated,
			bodyContains: `"correlation_id":"1"`,
		},
		{
			name:         "success multiple",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `[{"correlation_id":"a","original_url":"https://ya.ru"},{"correlation_id":"b","original_url":"https://yandex.ru"}]`,
			expectedCode: http.StatusCreated,
			bodyContains: `"correlation_id":"a"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/shorten/batch", strings.NewReader(tc.body))
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

func TestCreateShortURL_Conflict(t *testing.T) {
	handler := NewURLHandler(newTestService(t), "")

	r := chi.NewRouter()
	r.Post("/", handler.CreateShortURLHandler().ServeHTTP)

	send := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr
	}

	rr1 := send("https://example.com")
	assert.Equal(t, http.StatusCreated, rr1.Code)
	firstURL := rr1.Body.String()
	assert.NotEmpty(t, firstURL)

	rr2 := send("https://example.com")
	assert.Equal(t, http.StatusConflict, rr2.Code)
	assert.Equal(t, firstURL, rr2.Body.String(), "409 должен вернуть тот же short URL")
}

func TestCreateShortURLV2_Conflict(t *testing.T) {
	handler := NewURLHandler(newTestService(t), "")

	r := chi.NewRouter()
	r.Post("/api/shorten", handler.CreateShortURLHandlerV2().ServeHTTP)

	send := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr
	}

	rr1 := send(`{"url":"https://example.com"}`)
	assert.Equal(t, http.StatusCreated, rr1.Code)

	var resp1 model.Response
	require.NoError(t, json.Unmarshal(rr1.Body.Bytes(), &resp1))
	assert.NotEmpty(t, resp1.Result)

	rr2 := send(`{"url":"https://example.com"}`)
	assert.Equal(t, http.StatusConflict, rr2.Code)

	var resp2 model.Response
	require.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &resp2))
	assert.Equal(t, resp1.Result, resp2.Result, "409 должен вернуть тот же short URL")
}
