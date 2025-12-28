package cs

import (
	"log/slog"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var Module = fx.Module("lib/cs", // configuration source
	fx.Provide(
		fx.Annotate(newKeeper, fx.As(new(Keeper))),
	),
)

func newKeeper(l *slog.Logger) *keeperViper {
	viper := viper.New()
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("reference")
	viper.ReadInConfig()
	viper.SetConfigName("application")
	viper.MergeInConfig()
	t := slog.String("t", "keeperViper")
	return &keeperViper{viper, l.With(t)}
}
