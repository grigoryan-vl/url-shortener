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
	model "github.com/grigoryan-vl/URL-shortener/internal/model/request"
	"github.com/grigoryan-vl/URL-shortener/internal/model/storage"
)

type URLService interface {
	CreateShortURL(ctx context.Context, url string) (string, error)
	CreateShortURLBatch(ctx context.Context, urls []string) ([]string, error)
	GetURLByID(ctx context.Context, id string) (string, error)
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

		shortURL, err := h.svc.CreateShortURL(req.Context(), url)

		var responseURL string
		if h.baseURL != "" {
			base := strings.TrimRight(h.baseURL, "/")
			responseURL = base + "/" + shortURL
		} else {
			responseURL = "http://" + req.Host + "/" + shortURL
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

		shortURL, err := h.svc.CreateShortURL(req.Context(), url)

		var responseURL string
		if h.baseURL != "" {
			base := strings.TrimRight(h.baseURL, "/")
			responseURL = base + "/" + shortURL
		} else {
			responseURL = "http://" + req.Host + "/" + shortURL
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

		codes, err := h.svc.CreateShortURLBatch(req.Context(), urls)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		resp := make([]model.BatchResponse, len(batchReq))
		for i, r := range batchReq {
			var responseURL string
			if h.baseURL != "" {
				base := strings.TrimRight(h.baseURL, "/")
				responseURL = base + "/" + codes[i]
			} else {
				responseURL = "http://" + req.Host + "/" + codes[i]
			}
			resp[i] = model.BatchResponse{
				CorrelationID: r.CorrelationID,
				ShortURL:      responseURL,
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
