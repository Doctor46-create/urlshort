package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/Doctor46-create/urlshort/internal/handler"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/config"
)

type Handler struct {
	urlHandler handler.URLHandler
}

func NewHandler(srvc service.Shortener, cfg config.Config) *Handler {
	return &Handler{
		urlHandler: handler.NewURLHandler(srvc, cfg),
	}
}

func (h *Handler) InitRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.urlHandler.ShortenURL)
	r.Get("/{shortKey}", h.urlHandler.RedirectURL)
	return r
}
