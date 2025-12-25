package server

import (
	"net/http"

	"github.com/Doctor46-create/urlshort/internal/handler"
)

func (a *Application) initHTTPServer() {
	h := handler.NewHandler(
		a.service,
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

