package handler

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	model "github.com/grigoryan-vl/url-shortener/internal/model/request"
)

type UrlService interface {
	CreateShortUrl(url string) (string, error)
	GetUrlById(id string) (string, error)
}

type Handler struct {
	urlService UrlService
	baseUrl    string
}

func NewHandler(urlService UrlService, baseUrl string) *Handler {
	return &Handler{
		urlService: urlService,
		baseUrl:    baseUrl,
	}
}

func (handler *Handler) CreateShortUrlHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost) //дополнительно сообщим клиенту, какой метод все же доступен для этого эндпоинта
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

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

		shortUrl, err := handler.urlService.CreateShortUrl(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var responseURL string
		if handler.baseUrl != "" {
			// Нормализуем: убираем завершающие слэши, затем добавляем один
			base := strings.TrimRight(handler.baseUrl, "/")
			responseURL = base + "/" + shortUrl
		} else {
			responseURL = "http://" + req.Host + "/" + shortUrl
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		w.Write([]byte(responseURL))
	}

	return http.HandlerFunc(fn)
}

func (handler *Handler) CreateShortUrlHandlerV2() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost) //дополнительно сообщим клиенту, какой метод все же доступен для этого эндпоинта
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

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

		var urlRequestBody model.UrlRequest
		err = json.Unmarshal(body, &urlRequestBody)
		if err != nil {
			http.Error(w, "Failed to unmarshal body", http.StatusBadRequest)
			return
		}

		url := strings.TrimSpace(string(urlRequestBody.Url))
		if url == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		shortUrl, err := handler.urlService.CreateShortUrl(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var responseURL string
		if handler.baseUrl != "" {
			// Нормализуем: убираем завершающие слэши, затем добавляем один
			base := strings.TrimRight(handler.baseUrl, "/")
			responseURL = base + "/" + shortUrl
		} else {
			responseURL = "http://" + req.Host + "/" + shortUrl
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

func (handler *Handler) GetUrlByIdHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet) //дополнительно сообщим клиенту, какой метод все же доступен для этого эндпоинта
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		urlId := strings.TrimSpace(chi.URLParam(req, "id"))
		if urlId == "" {
			http.Error(w, "UrlId is required", http.StatusBadRequest)
			return
		}

		url, err := handler.urlService.GetUrlById(urlId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Redirect(w, req, url, http.StatusTemporaryRedirect)
	}

	return http.HandlerFunc(fn)
}
