package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestGetURLByIDHandler(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	existingCode, err := svc.CreateShortURL(context.Background(), "https://example.com", "test-user")
	if err != nil {
		t.Fatalf("failed to create short url: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Sync()
	sugar := *logger.Sugar()

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

func TestGetUserURLsHandler(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	authMw := middleware.Auth([]byte("test-secret"))

	r := chi.NewRouter()
	r.Post("/", authMw(handler.CreateShortURLHandler()).ServeHTTP)
	r.Get("/api/user/urls", authMw(handler.GetUserURLsHandler()).ServeHTTP)

	// 1. пустой список — 204 + кука в ответе
	req1 := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)

	assert.Equal(t, http.StatusNoContent, rr1.Code)
	cookies := rr1.Result().Cookies()
	require.NotEmpty(t, cookies, "auth middleware должна поставить куку")
	cookie := cookies[0]

	// 2. создаём URL с той же кукой
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req2.Header.Set("Content-Type", "text/plain")
	req2.AddCookie(cookie)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusCreated, rr2.Code)

	// 3. теперь список не пустой — 200 с JSON
	req3 := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req3.AddCookie(cookie)
	rr3 := httptest.NewRecorder()
	r.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)

	var urls []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	require.NoError(t, json.Unmarshal(rr3.Body.Bytes(), &urls))
	require.Len(t, urls, 1)
	assert.Equal(t, "https://example.com", urls[0].OriginalURL)
	assert.NotEmpty(t, urls[0].ShortURL)
	assert.Contains(t, urls[0].ShortURL, "http://")
}

func TestGetUserURLsHandler_Unauthorized(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	// без auth-миддлвари: userID в контексте пустой → 401
	r := chi.NewRouter()
	r.Get("/api/user/urls", handler.GetUserURLsHandler().ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestDeleteUserURLsHandler(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	authMw := middleware.Auth([]byte("test-secret"))

	r := chi.NewRouter()
	r.Post("/", authMw(handler.CreateShortURLHandler()).ServeHTTP)
	r.Get("/api/user/urls", authMw(handler.GetUserURLsHandler()).ServeHTTP)
	r.Delete("/api/user/urls", authMw(handler.DeleteUserURLsHandler()).ServeHTTP)
	r.Get("/{id}", handler.GetURLByIDHandler().ServeHTTP)

	// 1. получаем куку
	req1 := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)
	require.Equal(t, http.StatusNoContent, rr1.Code)
	cookie := rr1.Result().Cookies()[0]

	// 2. создаём два URL
	codes := make([]string, 0, 2)
	for _, u := range []string{"http://a.com", "http://b.com"} {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(u))
		req.Header.Set("Content-Type", "text/plain")
		req.AddCookie(cookie)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
		body := rr.Body.String()
		codes = append(codes, body[strings.LastIndex(body, "/")+1:])
	}

	// 3. проверяем, что оба работают
	for _, code := range codes {
		req := httptest.NewRequest(http.MethodGet, "/"+code, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	}

	// 4. удаляем один
	body, _ := json.Marshal([]string{codes[0]})
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(string(body)))
	reqDel.Header.Set("Content-Type", "application/json")
	reqDel.AddCookie(cookie)
	rrDel := httptest.NewRecorder()
	r.ServeHTTP(rrDel, reqDel)
	require.Equal(t, http.StatusAccepted, rrDel.Code)

	// 5. ждём, пока фоновый горутин проставит флаг (до 1 секунды)
	require.Eventually(t, func() bool {
		req := httptest.NewRequest(http.MethodGet, "/"+codes[0], nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr.Code == http.StatusGone
	}, time.Second, 10*time.Millisecond, "удалённый URL должен отдавать 410")

	// 6. второй URL не тронут
	req2 := httptest.NewRequest(http.MethodGet, "/"+codes[1], nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTemporaryRedirect, rr2.Code)

	// 7. в списке пользователя остался только один
	reqList := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	reqList.AddCookie(cookie)
	rrList := httptest.NewRecorder()
	r.ServeHTTP(rrList, reqList)
	require.Equal(t, http.StatusOK, rrList.Code)

	var urls []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	require.NoError(t, json.Unmarshal(rrList.Body.Bytes(), &urls))
	require.Len(t, urls, 1)
	assert.Equal(t, "http://b.com", urls[0].OriginalURL)
}

func TestDeleteUserURLsHandler_Validation(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	authMw := middleware.Auth([]byte("test-secret"))

	r := chi.NewRouter()
	r.Delete("/api/user/urls", authMw(handler.DeleteUserURLsHandler()).ServeHTTP)

	testCases := []struct {
		name         string
		contentType  string
		body         string
		withCookie   bool
		expectedCode int
	}{
		{
			name:         "wrong content type",
			contentType:  "text/plain",
			body:         `["abc"]`,
			withCookie:   true,
			expectedCode: http.StatusUnsupportedMediaType,
		},
		{
			name:         "empty body",
			contentType:  "application/json",
			body:         "",
			withCookie:   true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "malformed json",
			contentType:  "application/json",
			body:         `[`,
			withCookie:   true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty list",
			contentType:  "application/json",
			body:         `[]`,
			withCookie:   true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "success",
			contentType:  "application/json",
			body:         `["abc"]`,
			withCookie:   true,
			expectedCode: http.StatusAccepted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			if tc.withCookie {
				// получаем валидную куку через auth-миддлварь: делаем пустой запрос
				probe := httptest.NewRequest(http.MethodDelete, "/api/user/urls", nil)
				rrProbe := httptest.NewRecorder()
				r.ServeHTTP(rrProbe, probe)
				if cookies := rrProbe.Result().Cookies(); len(cookies) > 0 {
					req.AddCookie(cookies[0])
				}
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}

func TestDeleteUserURLsHandler_Unauthorized(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	// без auth-миддлвари: userID в контексте пустой → 401
	r := chi.NewRouter()
	r.Delete("/api/user/urls", handler.DeleteUserURLsHandler().ServeHTTP)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["abc"]`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestDeleteUserURLsHandler_NotOwner(t *testing.T) {
	svc := newTestService(t)
	handler := NewURLHandler(svc, "")

	authMw := middleware.Auth([]byte("test-secret"))

	r := chi.NewRouter()
	r.Post("/", authMw(handler.CreateShortURLHandler()).ServeHTTP)
	r.Delete("/api/user/urls", authMw(handler.DeleteUserURLsHandler()).ServeHTTP)
	r.Get("/{id}", handler.GetURLByIDHandler().ServeHTTP)

	// пользователь 1 создаёт URL
	reqProbe1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://a.com"))
	reqProbe1.Header.Set("Content-Type", "text/plain")
	rrProbe1 := httptest.NewRecorder()
	r.ServeHTTP(rrProbe1, reqProbe1)
	cookie1 := rrProbe1.Result().Cookies()[0]
	body1 := rrProbe1.Body.String()
	code := body1[strings.LastIndex(body1, "/")+1:]

	// пользователь 2 пытается удалить
	reqProbe2 := httptest.NewRequest(http.MethodDelete, "/api/user/urls", nil)
	rrProbe2 := httptest.NewRecorder()
	r.ServeHTTP(rrProbe2, reqProbe2)
	cookie2 := rrProbe2.Result().Cookies()[0]

	body, _ := json.Marshal([]string{code})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie2)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusAccepted, rr.Code)

	// URL не должен быть удалён
	time.Sleep(100 * time.Millisecond) // даём фоновой горутине отработать
	reqGet := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	rrGet := httptest.NewRecorder()
	r.ServeHTTP(rrGet, reqGet)
	assert.Equal(t, http.StatusTemporaryRedirect, rrGet.Code, "чужой URL не должен удаляться")

	_ = cookie1 // чтобы не было unused
}
