package kv

import (
	"log/slog"
	"reflect"

	"github.com/spf13/viper"
)

func newViperDriver(logger *slog.Logger) *viperDriver {
	viper := viper.New()
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("reference")
	viper.ReadInConfig()
	viper.SetConfigName("application")
	viper.MergeInConfig()
	name := slog.String("name", reflect.TypeFor[viperDriver]().Name())
	return &viperDriver{viper, logger.With(name)}
}

type viperDriver struct {
	viper *viper.Viper
	log   *slog.Logger
}

func (d *viperDriver) Load(key string, dst any) error {
	err := d.viper.UnmarshalKey(key, dst)
	if err != nil {
		d.log.Error("load failed", slog.String("key", key), slog.Any("reason", err))
		return err
	}
	d.log.Info("load succeed", slog.String("key", key), slog.Any("val", dst))
	return nil
}
