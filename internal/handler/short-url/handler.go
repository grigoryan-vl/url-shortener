package handler

import (
	"io"
	"mime"
	"net/http"
	"strings"

	service "github.com/grigoryan-vl/url-shortener/internal/service/short-url"
)

type Handler struct {
	urlService *service.UrlService
}

func NewHandler(urlService *service.UrlService) *Handler {
	return &Handler{
		urlService: urlService,
	}
}

func (handler *Handler) CreateShortUrlHandler(w http.ResponseWriter, req *http.Request) {
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

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(shortUrl))
}

func (handler *Handler) GetUrlByIdHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet) //дополнительно сообщим клиенту, какой метод все же доступен для этого эндпоинта
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	urlIdFromPath := strings.TrimPrefix(req.URL.Path, "/")

	urlId := strings.TrimSpace(urlIdFromPath)
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
