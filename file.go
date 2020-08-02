package fido

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FilesProviderName is the name of the FilesProvider.
const FilesProviderName = "Files"

// FromFiles returns a FileProvider wrapping the given ReadProvider. The given patterns should absolute
// paths or globs.
func FromFiles(provider ReadProvider, patterns ...string) *FileProvider {
	return &FileProvider{
		patterns: patterns,
		matches:  make(map[string]struct{}),
		provider: provider,
		open: func(name string) (io.ReadCloser, error) {
			return os.Open(name)
		},
	}
}

// FileProvider provides a standard Provider that wraps a given ReadProvider.
type FileProvider struct {
	patterns []string
	matches  map[string]struct{}
	provider ReadProvider
	open     func(string) (io.ReadCloser, error)
}

func (p *FileProvider) String() string {
	return JoinProviderNames(p.provider.String(), FilesProviderName)
}

// Values searches for files matching the patterns provided, opening each file and passing them to
// the given ReadProvider for processing.
func (p *FileProvider) Values(ctx context.Context, callback Callback) error {
	for _, pattern := range p.patterns {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("%w: failed to pattern match on %s", err, pattern)
			}

			for _, path := range matches {
				if _, ok := p.matches[path]; ok {
					continue
				}

				f, err := p.open(path)
				if err != nil {
					return fmt.Errorf("%w: failed to open %s", err, path)
				}

				p.matches[path] = struct{}{}

				if err := p.provider.Values(ctx, f, callback); err != nil {
					return err
				}

				if err := f.Close(); err != nil {
					return fmt.Errorf("%w: failed to close %s", err, path)
				}
			}
		}
	}

	return nil
}
