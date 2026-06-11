package metrics

import "time"

type DatabaseActivity struct {
	DatabaseName string

	ProcessID int

	UserName        string
	ApplicationName string
	ClientAddress   string

	State string
	Query string

	BackendStartedAt time.Time
	QueryStartedAt   *time.Time
	StateChangedAt   *time.Time
}

type DatabaseActivitySnapshot struct {
	CollectedAt time.Time

	DatabaseInstanceID string

	Activities []DatabaseActivity
}
