package core

import (
	"time"

	"github.com/nullsecurity-australia/mariadb-backup/pkg/storage"
)

type PruneOptions struct {
	Targets   []storage.Storage
	Retention string
	Now       time.Time
}
