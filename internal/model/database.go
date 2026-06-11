package model

type DatabaseInfo struct {
	Version string
}

type DatabaseStats struct {
	DatabaseName string
	NumBackends  int
	XactCommit   int64
	XactRollback int64
	BlksRead     int64
	BlksHit      int64
	TupReturned  int64
	TupFetched   int64
	TupInserted  int64
	TupUpdated   int64
	TupDeleted   int64
}
