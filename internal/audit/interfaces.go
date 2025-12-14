package audit

type Observer interface {
	Notify(event *AuditEvent) error
	Close() error
}
