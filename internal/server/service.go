package server

import "github.com/Doctor46-create/urlshort/internal/service"

func (a *Application) initService() {
	a.service = service.NewURLService(a.repo)
}
