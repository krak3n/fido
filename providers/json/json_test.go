package json

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

type ReaderFunc func(p []byte) (n int, err error)

func (fn ReaderFunc) Read(p []byte) (int, error) {
	return fn(p)
}

func TestProviderErrors(t *testing.T) {
	cases := map[string]struct {
		reader io.Reader
		err    func(*testing.T, error)
	}{
		"ReaderError": {
			reader: ReaderFunc(func([]byte) (int, error) {
				return 0, io.ErrClosedPipe
			}),
			err: func(t *testing.T, err error) {
				want := io.ErrClosedPipe

				if !errors.Is(err, want) {
					t.Errorf("want %+v, got %+v", want, err)
				}
			},
		},
		"UnmarshalError": {
			reader: strings.NewReader("invalid json"),
			err: func(t *testing.T, err error) {
				t.Log(err)

				if _, ok := errors.Unwrap(err).(*json.SyntaxError); !ok {
					t.Errorf("want *json.SyntaxError, got %+v", err)
				}
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			provider := New()

			tc.err(t, provider.Values(context.Background(), tc.reader, nil))
		})
	}
}
