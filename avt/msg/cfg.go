package msg

type props struct {
	Protocol protocol `mapstructure:"protocol"`
	Mapping  mapping  `mapstructure:"mapping"`
}

type protocol struct {
	Modes []string `mapstructure:"modes"`
	Http  http     `mapstructure:"http"`
}

type mapping struct {
	Modes []string `mapstructure:"modes"`
}

type http struct {
	Port int `mapstructure:"port"`
}
