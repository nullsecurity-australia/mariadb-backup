package core

import (
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/nullsecurity-australia/mariadb-backup/pkg/storage"
	"github.com/nullsecurity-australia/mariadb-backup/pkg/storage/credentials"
	"github.com/stretchr/testify/assert"
)

func TestConvertToHours(t *testing.T) {
	tests := []struct {
		input  string
		output int
		err    error
	}{
		{"2h", 2, nil},
		{"3w", 3 * 7 * 24, nil},
		{"5d", 5 * 24, nil},
		{"1m", 30 * 24, nil},
		{"1y", 365 * 24, nil},
		{"100x", 0, fmt.Errorf("invalid format: 100x")},
	}
	for _, tt := range tests {
		hours, err := convertToHours(tt.input)
		switch {
		case (err == nil && tt.err != nil) || (err != nil && tt.err == nil):
			t.Errorf("expected error %v, got %v", tt.err, err)
		case err != nil && tt.err != nil && err.Error() != tt.err.Error():
			t.Errorf("expected error %v, got %v", tt.err, err)
		case hours != tt.output:
			t.Errorf("input %s expected %d, got %d", tt.input, tt.output, hours)
		}
	}
}

func TestPrune(t *testing.T) {
	// we use a fixed list of file before, and a subset of them for after
	// db_backup_YYYY-MM-DDTHH:mm:ssZ.<compression>
	// our list of timestamps should give us these files, of the following time ago:
	// 0.25h, 1h, 2h, 3h, 24h (1d), 36h (1.5d), 48h (2d), 60h (2.5d) 72h(3d),
	// 167h (1w-1h), 168h (1w), 240h (1.5w) 336h (2w), 576h (2.5w), 504h (3w)
	// 744h (3.5w), 720h (1m), 1000h (1.5m), 1440h (2m), 1800h (2.5m), 2160h (3m),
	// 8760h (1y), 12000h (1.5y), 17520h (2y)
	// we use a fixed starting time to make it consistent.
	now := time.Date(2021, 1, 1, 0, 30, 0, 0, time.UTC)
	hoursAgo := []float32{0.25, 1, 2, 3, 24, 36, 48, 60, 72, 167, 168, 240, 336, 504, 576, 744, 720, 1000, 1440, 1800, 2160, 8760, 12000, 17520}
	// convert to filenames
	var filenames []string
	for _, h := range hoursAgo {
		// convert the time diff into a duration, do not forget the negative
		duration, err := time.ParseDuration(fmt.Sprintf("-%fh", h))
		if err != nil {
			t.Fatalf("failed to parse duration: %v", err)
		}
		// convert it into a time.Time
		// and add 30 mins to our "now" time.
		relativeTime := now.Add(duration).Add(-30 * time.Minute)
		// convert that into the filename
		filename := fmt.Sprintf("db_backup_%sZ.gz", relativeTime.Format("2006-01-02T15:04:05"))
		filenames = append(filenames, filename)
	}
	tests := []struct {
		name        string
		opts        PruneOptions
		beforeFiles []string
		afterFiles  []string
		err         error
	}{
		{"invalid format", PruneOptions{Retention: "100x", Now: now}, nil, nil, fmt.Errorf("invalid retention string: 100x")},
		{"no targets", PruneOptions{Retention: "1h", Now: now}, nil, nil, fmt.Errorf("no targets")},
		// 1 hour - file[1] is 1h+30m = 1.5h, so it should be pruned
		{"1 hour", PruneOptions{Retention: "1h", Now: now}, filenames, filenames[0:1], nil},
		// 2 hours - file[2] is 2h+30m = 2.5h, so it should be pruned
		{"2 hours", PruneOptions{Retention: "2h", Now: now}, filenames, filenames[0:2], nil},
		// 2 days - file[6] is 48h+30m = 48.5h, so it should be pruned
		{"2 days", PruneOptions{Retention: "2d", Now: now}, filenames, filenames[0:6], nil},
		// 3 weeks - file[13] is 504h+30m = 504.5h, so it should be pruned
		{"3 weeks", PruneOptions{Retention: "3w", Now: now}, filenames, filenames[0:13], nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a temporary directory
			workDir := t.TempDir()
			// create beforeFiles in the directory and create a target, but only if there are beforeFiles
			// this lets us also test no targets, which should generate an error
			if len(tt.beforeFiles) > 0 {
				for _, filename := range tt.beforeFiles {
					if err := os.WriteFile(fmt.Sprintf("%s/%s", workDir, filename), nil, 0644); err != nil {
						t.Errorf("failed to create file %s: %v", filename, err)
						return
					}
				}

				// add our tempdir as the target
				store, err := storage.ParseURL(fmt.Sprintf("file://%s", workDir), credentials.Creds{})
				if err != nil {
					t.Errorf("failed to parse url: %v", err)
					return
				}

				tt.opts.Targets = append(tt.opts.Targets, store)
			}

			// run Prune
			err := Prune(tt.opts)
			switch {
			case (err == nil && tt.err != nil) || (err != nil && tt.err == nil):
				t.Errorf("expected error %v, got %v", tt.err, err)
			case err != nil && tt.err != nil && err.Error() != tt.err.Error():
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
			// check files match
			files, err := os.ReadDir(workDir)
			if err != nil {
				t.Errorf("failed to read directory: %v", err)
				return
			}
			var afterFiles []string
			for _, file := range files {
				afterFiles = append(afterFiles, file.Name())
			}
			slices.Sort(afterFiles)
			slices.Sort(tt.afterFiles)
			assert.ElementsMatch(t, tt.afterFiles, afterFiles)
		})
	}
}
