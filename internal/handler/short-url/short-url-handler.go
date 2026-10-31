package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	middleware "github.com/grigoryan-vl/URL-shortener/internal/middleware"
	model "github.com/grigoryan-vl/URL-shortener/internal/model/request"
	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
)

type URLService interface {
	CreateShortURL(ctx context.Context, url, userID string) (string, error)
	CreateShortURLBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	GetURLByID(ctx context.Context, id string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]storage.UserURL, error)
}

type URLHandler struct {
	svc     URLService
	baseURL string
}

func NewURLHandler(svc URLService, baseURL string) *URLHandler {
	return &URLHandler{
		svc:     svc,
		baseURL: baseURL,
	}
}

func (h *URLHandler) buildURL(req *http.Request, code string) string {
	if h.baseURL != "" {
		return strings.TrimRight(h.baseURL, "/") + "/" + code
	}
	return "http://" + req.Host + "/" + code
}

func (h *URLHandler) CreateShortURLHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "text/plain" {
			http.Error(w, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		url := strings.TrimSpace(string(body))
		if url == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		userID := middleware.UserIDFromContext(req.Context())
		shortURL, err := h.svc.CreateShortURL(req.Context(), url, userID)

		var responseURL string
		if shortURL != "" {
			responseURL = h.buildURL(req, shortURL)
		}

		if errors.Is(err, storage.ErrConflict) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(responseURL))
			return
		}
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseURL))
	}

	return http.HandlerFunc(fn)
}

func (h *URLHandler) CreateShortURLHandlerV2() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var urlReq model.URLRequest
		if err := json.Unmarshal(body, &urlReq); err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusBadRequest)
			return
		}

		url := strings.TrimSpace(urlReq.URL)
		if url == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		userID := middleware.UserIDFromContext(req.Context())
		shortURL, err := h.svc.CreateShortURL(req.Context(), url, userID)

		var responseURL string
		if shortURL != "" {
			responseURL = h.buildURL(req, shortURL)
		}

		if errors.Is(err, storage.ErrConflict) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			result, mErr := json.Marshal(model.Response{Result: responseURL})
			if mErr != nil {
				http.Error(w, mErr.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(result)
			return
		}
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		result, err := json.Marshal(model.Response{Result: responseURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(result)
	}

	return http.HandlerFunc(fn)
}

func (h *URLHandler) CreateShortURLBatchHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var batchReq []model.BatchRequest
		if err := json.Unmarshal(body, &batchReq); err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusBadRequest)
			return
		}
		if len(batchReq) == 0 {
			http.Error(w, "Empty batch", http.StatusBadRequest)
			return
		}

		urls := make([]string, len(batchReq))
		for i, r := range batchReq {
			urls[i] = strings.TrimSpace(r.OriginalURL)
			if urls[i] == "" {
				http.Error(w, "URL is required", http.StatusBadRequest)
				return
			}
		}

		userID := middleware.UserIDFromContext(req.Context())
		codes, err := h.svc.CreateShortURLBatch(req.Context(), urls, userID)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		resp := make([]model.BatchResponse, len(batchReq))
		for i, r := range batchReq {
			resp[i] = model.BatchResponse{
				CorrelationID: r.CorrelationID,
				ShortURL:      h.buildURL(req, codes[i]),
			}
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(buf.Bytes())
	}

	return http.HandlerFunc(fn)
}

func (h *URLHandler) GetURLByIDHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		id := strings.TrimSpace(chi.URLParam(req, "id"))
		if id == "" {
			http.Error(w, "URLId is required", http.StatusBadRequest)
			return
		}

		url, err := h.svc.GetURLByID(req.Context(), id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				http.Error(w, "URL not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, url, http.StatusTemporaryRedirect)
	}

	return http.HandlerFunc(fn)
}

func (h *URLHandler) GetUserURLsHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		userID := middleware.UserIDFromContext(req.Context())
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		urls, err := h.svc.GetUserURLs(req.Context(), userID)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		type item struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}

		resp := make([]item, len(urls))
		for i, u := range urls {
			resp[i] = item{
				ShortURL:    h.buildURL(req, u.Short),
				OriginalURL: u.Original,
			}
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}

	return http.HandlerFunc(fn)
}
