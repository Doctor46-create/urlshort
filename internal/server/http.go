package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/handler"
	"github.com/Doctor46-create/urlshort/internal/service"
)

func (a *Application) initHTTPServer() {
	srvc := service.NewURLService(a.repo)

	h := handler.NewHandler(
		srvc,
		a.cfg,
		a.log,
		a.dbConfig,
		a.audit,
	)

	a.httpServer = &http.Server{
		Addr:    a.cfg.GetAddress(),
		Handler: h.InitRouter(a.log, *a.cfg),
	}
}
