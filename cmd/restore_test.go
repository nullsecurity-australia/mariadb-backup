package cmd

import (
	"net/url"
	"testing"

	"github.com/nullsecurity-australia/mariadb-backup/pkg/compression"
	"github.com/nullsecurity-australia/mariadb-backup/pkg/database"
	"github.com/nullsecurity-australia/mariadb-backup/pkg/storage"
	"github.com/nullsecurity-australia/mariadb-backup/pkg/storage/file"
)

func TestRestoreCmd(t *testing.T) {
	t.Parallel()

	fileTarget := "file:///foo/bar"
	fileTargetURL, _ := url.Parse(fileTarget)

	tests := []struct {
		name                 string
		args                 []string // "restore" will be prepended automatically
		config               string
		wantErr              bool
		expectedTarget       storage.Storage
		expectedTargetFile   string
		expectedDbconn       database.Connection
		expectedDatabasesMap map[string]string
		expectedCompressor   compression.Compressor
	}{
		{"missing server and target options", []string{""}, "", true, nil, "", database.Connection{}, nil, &compression.GzipCompressor{}},
		{"invalid target URL", []string{"--server", "abc", "--target", "def"}, "", true, nil, "", database.Connection{Host: "abc"}, nil, &compression.GzipCompressor{}},
		{"valid URL missing dump filename", []string{"--server", "abc", "--target", "file:///foo/bar"}, "", true, nil, "", database.Connection{Host: "abc"}, nil, &compression.GzipCompressor{}},
		{"valid file URL", []string{"--server", "abc", "--target", fileTarget, "filename.tgz", "--verbose", "2"}, "", false, file.New(*fileTargetURL), "filename.tgz", database.Connection{Host: "abc", Port: defaultPort}, map[string]string{}, &compression.GzipCompressor{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockExecs()
			m.On("restore", tt.expectedTarget, tt.expectedTargetFile, tt.expectedDbconn, tt.expectedDatabasesMap, tt.expectedCompressor).Return(nil)
			cmd, err := rootCmd(m)
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs(append([]string{"restore"}, tt.args...))
			err = cmd.Execute()
			switch {
			case err == nil && tt.wantErr:
				t.Fatal("missing error")
			case err != nil && !tt.wantErr:
				t.Fatal(err)
			case err == nil:
				m.AssertExpectations(t)
			}

		})
	}
}
