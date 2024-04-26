package config

import (
	"github.com/nullsecurity-australia/mariadb-backup/pkg/remote"
)

type RemoteSpec struct {
	remote.Connection
}
