//+build go1.14,!windows

package fido

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type ReadCloser struct {
	ReadFn  func([]byte) (int, error)
	CloseFn func() error
}

func (rc *ReadCloser) Read(b []byte) (int, error) {
	if rc.ReadFn != nil {
		return rc.ReadFn(b)
	}

	return 0, nil
}

func (rc *ReadCloser) Close() error {
	if rc.CloseFn != nil {
		return rc.CloseFn()
	}

	return nil
}

// NOTE: this test is only run for go1.14+ due to the test cleanup method
func TestFilesProvider(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), t.Name())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})

	t1, err := ioutil.TempFile(dir, "t1.cfg")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := t1.Write([]byte("foo:bar")); err != nil {
		t.Fatal(err)
	}

	t.Log("created:", t1.Name())

	cases := map[string]struct {
		ctx      context.Context
		provider ReadProvider
		patterns []string
		callback func(*testing.T) Callback
		openfn   func(string) (io.ReadCloser, error)
		err      error
	}{
		"ContextCanceled": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				return ctx
			}(),
			patterns: []string{t1.Name()},
			callback: func(t *testing.T) Callback {
				return func(Path, interface{}) error {
					return nil
				}
			},
			err: context.Canceled,
		},
		"BadPattern": {
			ctx:      context.Background(),
			patterns: []string{"[]a]"},
			callback: func(t *testing.T) Callback {
				return func(Path, interface{}) error {
					return nil
				}
			},
			err: filepath.ErrBadPattern,
		},
		"OpenError": {
			ctx: context.Background(),
			patterns: []string{
				os.DevNull,
			},
			openfn: func(string) (io.ReadCloser, error) {
				return nil, os.ErrNotExist
			},
			callback: func(t *testing.T) Callback {
				return func(Path, interface{}) error {
					return nil
				}
			},
			err: os.ErrNotExist,
		},
		"ValuesError": {
			ctx: context.Background(),
			patterns: []string{
				t1.Name(),
			},
			provider: NewTestReadProvider(t, func(ctx context.Context, reader io.Reader, callback Callback) error {
				return ErrSetInvalidValue
			}),
			callback: func(t *testing.T) Callback {
				return func(Path, interface{}) error {
					return nil
				}
			},
			err: ErrSetInvalidValue,
		},
		"CloseError": {
			ctx: context.Background(),
			patterns: []string{
				os.DevNull,
			},
			openfn: func(string) (io.ReadCloser, error) {
				return &ReadCloser{
					CloseFn: func() error {
						return os.ErrClosed
					},
				}, nil
			},
			provider: NewTestReadProvider(t, func(ctx context.Context, reader io.Reader, callback Callback) error {
				return nil
			}),
			callback: func(t *testing.T) Callback {
				return func(Path, interface{}) error {
					return nil
				}
			},
			err: os.ErrClosed,
		},
		"ReadsFile": {
			ctx: context.Background(),
			patterns: []string{
				t1.Name(),
				t1.Name(), // Prevents duplicate
			},
			provider: NewTestReadProvider(t, func(ctx context.Context, reader io.Reader, callback Callback) error {
				b, err := ioutil.ReadAll(reader)
				if err != nil {
					return err
				}

				parts := strings.Split(string(b), ":")

				return callback(Path{parts[0]}, parts[1])
			}),
			callback: func(t *testing.T) Callback {
				return func(path Path, value interface{}) error {
					{
						want := Path{"foo"}
						if !reflect.DeepEqual(path, want) {
							t.Errorf("want %+v, got %+v", want, path)
						}
					}

					{
						want := "bar"
						if !reflect.DeepEqual(value, want) {
							t.Errorf("want %+v, got %+v", want, value)
						}
					}

					return nil
				}
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := FromFiles(tc.provider, tc.patterns...)
			if tc.openfn != nil {
				p.open = tc.openfn
			}

			err := p.Values(tc.ctx, tc.callback(t))
			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", tc.err, err)
			}
		})
	}
}
