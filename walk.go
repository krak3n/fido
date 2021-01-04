package fido

import (
	"context"
	"reflect"
)

// WalkMap traverses the given map calling the provided callback function.
func WalkMap(ctx context.Context, src map[string]interface{}, path Path, writer Writer) error {
	for key, value := range src {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			rv := reflect.ValueOf(value)

			if rv.Kind() == reflect.Map {
				if err := WalkMap(ctx, value.(map[string]interface{}), append(path, key), writer); err != nil {
					return err
				}

				continue
			}

			if err := writer.Write(append(path, key), value); err != nil {
				return err
			}
		}
	}

	return nil
}
