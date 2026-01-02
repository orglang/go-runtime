package ws

import (
	"orglang/orglang/lib/ck"
)

func newExchangePC(k ck.Loader) (exchangePC, error) {
	pc := &exchangePC{}
	loadingErr := k.Load("exchange", pc)
	if loadingErr != nil {
		return exchangePC{}, loadingErr
	}
	validationErr := pc.Validate()
	if validationErr != nil {
		return exchangePC{}, validationErr
	}
	return *pc, nil
}

type exchangePC struct {
	Protocol protocolPC `mapstructure:"protocol"`
	Server   serverPC   `mapstructure:"server"`
}

type protocolPC struct {
	Modes []protocolMode `mapstructure:"modes"`
	Http  httpPC         `mapstructure:"http"`
}

type serverPC struct {
	Mode serverMode `mapstructure:"mode"`
	Echo echoPC     `mapstructure:"echo"`
}

type httpPC struct {
	Port int `mapstructure:"port"`
}

type echoPC struct{}

type protocolMode string

const (
	httpMode = protocolMode("http")
)

type serverMode string

const (
	echoMode = serverMode("echo")
)
