package core

import (
	"github.com/nullsecurity-australia/mariadb-backup/pkg/compression"
	"github.com/nullsecurity-australia/mariadb-backup/pkg/database"
	"github.com/nullsecurity-australia/mariadb-backup/pkg/storage"
)

type DumpOptions struct {
	Targets             []storage.Storage
	Safechars           bool
	DBNames             []string
	DBConn              database.Connection
	Compressor          compression.Compressor
	Exclude             []string
	PreBackupScripts    string
	PostBackupScripts   string
	Compact             bool
	SuppressUseDatabase bool
	MaxAllowedPacket    int
}
