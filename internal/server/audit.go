package server

import (
	"github.com/Doctor46-create/urlshort/internal/audit"
	"go.uber.org/zap"
)

func (a *Application) initAudit() {
	if !a.cfg.HasAudit() {
		a.log.Info("Audit disabled")
		return
	}

	subject := audit.NewSubject()

	if path := a.cfg.GetAuditFile(); path != "" {
		observer, err := audit.NewFileObserver(path)
		if err != nil {
			a.log.Error("Failed to init file audit", zap.Error(err))
		} else {
			subject.Register(observer)
			a.log.Info("File audit enabled", zap.String("path", path))
		}
	}

	if url := a.cfg.GetAuditURL(); url != "" {
		subject.Register(audit.NewHTTPObserver(url))
		a.log.Info("HTTP audit enabled", zap.String("url", url))
	}

	a.audit = subject
}
