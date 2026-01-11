package property

import (
	"log/slog"

	"github.com/spf13/viper"
)

const (
	key = "exchange"
)

type viperPAO struct {
	viper *viper.Viper
	log   *slog.Logger
}

func newViperPAO() *viperPAO {
	return &viperPAO{}
}

func (pao *viperPAO) Load() (ExchangeCS, error) {
	val := &ExchangeCS{}
	err := pao.viper.UnmarshalKey(key, val)
	if err != nil {
		pao.log.Error("load failed", slog.String("key", key), slog.Any("reason", err))
		return *val, err
	}
	pao.log.Info("load succeed", slog.String("key", key), slog.Any("val", val))
	return *val, nil
}
