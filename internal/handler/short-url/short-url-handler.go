package handler

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	model "github.com/grigoryan-vl/URL-shortener/internal/model/request"
)

type URLService interface {
	CreateShortURL(URL string) (string, error)
	GetURLByID(id string) (string, error)
}

type Handler struct {
	URLService URLService
	baseURL    string
}

func NewHandler(URLService URLService, baseURL string) *Handler {
	return &Handler{
		URLService: URLService,
		baseURL:    baseURL,
	}
}

func (handler *Handler) CreateShortURLHandler() http.Handler {
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

		URL := strings.TrimSpace(string(body))
		if URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		shortURL, err := handler.URLService.CreateShortURL(URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var responseURL string
		if handler.baseURL != "" {
			// Нормализуем: убираем завершающие слэши, затем добавляем один
			base := strings.TrimRight(handler.baseURL, "/")
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

func (handler *Handler) CreateShortURLHandlerV2() http.Handler {
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

		var URLRequestBody model.URLRequest
		err = json.Unmarshal(body, &URLRequestBody)
		if err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusBadRequest)
			return
		}

		URL := strings.TrimSpace(string(URLRequestBody.URL))
		if URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		shortURL, err := handler.URLService.CreateShortURL(URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var responseURL string
		if handler.baseURL != "" {
			// Нормализуем: убираем завершающие слэши, затем добавляем один
			base := strings.TrimRight(handler.baseURL, "/")
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

func (handler *Handler) GetURLByIDHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		URLId := strings.TrimSpace(chi.URLParam(req, "id"))
		if URLId == "" {
			http.Error(w, "URLId is required", http.StatusBadRequest)
			return
		}

		URL, err := handler.URLService.GetURLByID(URLId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Redirect(w, req, URL, http.StatusTemporaryRedirect)
	}

	return http.HandlerFunc(fn)
}
