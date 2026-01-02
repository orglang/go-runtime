package ck

type Loader interface {
	Load(key string, val any) error
}
