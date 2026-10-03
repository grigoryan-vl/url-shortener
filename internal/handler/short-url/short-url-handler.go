package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	model "github.com/grigoryan-vl/URL-shortener/internal/model/request"
	service "github.com/grigoryan-vl/URL-shortener/internal/service/short-url"
)

type URLService interface {
	CreateShortURL(URL string) (string, error)
	GetURLByID(id string) (string, error)
}

type Handler struct {
	svc     URLService
	baseURL string
}

func NewHandler(svc URLService, baseURL string) *Handler {
	return &Handler{
		svc:     svc,
		baseURL: baseURL,
	}
}

func (h *Handler) CreateShortURLHandler() http.Handler {
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

		shortURL, err := h.svc.CreateShortURL(url)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		var responseURL string
		if h.baseURL != "" {
			base := strings.TrimRight(h.baseURL, "/")
			responseURL = base + "/" + shortURL
		} else {
			responseURL = "http://" + req.Host + "/" + shortURL
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseURL))
	}

	return http.HandlerFunc(fn)
}

func (h *Handler) CreateShortURLHandlerV2() http.Handler {
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

		shortURL, err := h.svc.CreateShortURL(url)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		var responseURL string
		if h.baseURL != "" {
			base := strings.TrimRight(h.baseURL, "/")
			responseURL = base + "/" + shortURL
		} else {
			responseURL = "http://" + req.Host + "/" + shortURL
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

func (h *Handler) GetURLByIDHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		id := strings.TrimSpace(chi.URLParam(req, "id"))
		if id == "" {
			http.Error(w, "URLId is required", http.StatusBadRequest)
			return
		}

		url, err := h.svc.GetURLByID(id)
		if err != nil {
			if errors.Is(err, service.ErrURLNotFound) {
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
