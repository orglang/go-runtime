package ck

import (
	"log/slog"
	"reflect"

	"github.com/spf13/viper"
)

func newKeeperViper(l *slog.Logger) *keeperViper {
	viper := viper.New()
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("reference")
	viper.ReadInConfig()
	viper.SetConfigName("application")
	viper.MergeInConfig()
	name := slog.String("name", reflect.TypeFor[keeperViper]().Name())
	return &keeperViper{viper, l.With(name)}
}

type keeperViper struct {
	viper *viper.Viper
	log   *slog.Logger
}

func (k *keeperViper) Load(key string, val any) error {
	err := k.viper.UnmarshalKey(key, val)
	if err != nil {
		k.log.Error("load failed", slog.String("key", key), slog.Any("reason", err))
		return err
	}
	k.log.Info("load succeed", slog.String("key", key), slog.Any("val", val))
	return nil
}
