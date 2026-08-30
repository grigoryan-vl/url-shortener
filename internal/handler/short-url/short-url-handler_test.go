package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	urlService "github.com/grigoryan-vl/url-shortener/internal/service/short-url"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortUrl(t *testing.T) {
	srv := urlService.NewUrlService()
	handler := NewHandler(srv, "")

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
			body:         `{"url":"https://example.com"}`,
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

			handler.CreateShortUrlHandler(rr, req)

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

func TestGetUrlByIdHandler(t *testing.T) {
	srv := urlService.NewUrlService()
	handler := NewHandler(srv, "")

	existingCode, _ := srv.CreateShortUrl("https://example.com")

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
			expectedCode: http.StatusBadRequest,
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

			handler.GetUrlByIdHandler(rr, req)

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

/*POST / HTTP/1.1
Host: localhost:8080
Content-Type: text/plain

https://practicum.yandex.ru/


HTTP/1.1 201 Created
Content-Type: text/plain
Content-Length: 30

http://localhost:8080/EwHXdJfB

----------------------------------
GET /EwHXdJfB HTTP/1.1
Host: localhost:8080
Content-Type: text/plain

*/
