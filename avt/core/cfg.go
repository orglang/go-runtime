package core

import (
	"log/slog"

	"github.com/spf13/viper"
)

// port
type Keeper interface {
	Load(key string, v any) error
}

// adapter
type keeperViper struct {
	viper  *viper.Viper
	logger *slog.Logger
}

func (k *keeperViper) Load(key string, v any) error {
	err := k.viper.UnmarshalKey(key, v)
	if err != nil {
		k.logger.Error("load failed", slog.String("key", key), slog.Any("reason", err))
		return err
	}
	k.logger.Info("load succeeded", slog.String("key", key), slog.Any("val", v))
	return nil
}
