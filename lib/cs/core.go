package cs

type Keeper interface {
	Load(key string, val any) error
}
