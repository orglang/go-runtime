package kv

import (
	"context"
)

type Loader interface {
	Load(key string, dst any) error
}

type Getter[T any] interface {
	Get(ctx context.Context, key string) (T, error)
}
