package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/config"
)

type Handler struct {
	urlHandler URLHandler
}

func NewHandler(srvc service.Shortener, cfg config.ServiceConfig) *Handler {
	return &Handler{
		urlHandler: NewURLHandler(srvc, cfg),
	}
}

func (h *Handler) InitRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.urlHandler.ShortenURL)
	r.Get("/{shortKey}", h.urlHandler.RedirectURL)
	return r
}
