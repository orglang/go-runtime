package cs

type Keeper interface {
	Load(key string, v any) error
}
