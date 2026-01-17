package ws

import (
	"orglang/go-runtime/lib/kv"
)

func newExchangeCS(loader kv.Loader) (exchangeCS, error) {
	dto := &exchangeCS{}
	loadingErr := loader.Load("exchange", dto)
	if loadingErr != nil {
		return exchangeCS{}, loadingErr
	}
	validationErr := dto.Validate()
	if validationErr != nil {
		return exchangeCS{}, validationErr
	}
	return *dto, nil
}

type exchangeCS struct {
	Protocol protocolCS `mapstructure:"protocol"`
	Server   serverCS   `mapstructure:"server"`
}

type protocolCS struct {
	Modes []protoModeCS `mapstructure:"modes"`
	Http  httpCS        `mapstructure:"http"`
}

type serverCS struct {
	Mode serverModeCS `mapstructure:"mode"`
	Echo echoCS       `mapstructure:"echo"`
}

type httpCS struct {
	Port uint16 `mapstructure:"port"`
}

type echoCS struct{}

type protoModeCS string

const (
	httpProto = protoModeCS("http")
)

type serverModeCS string

const (
	echoServer = serverModeCS("echo")
)
