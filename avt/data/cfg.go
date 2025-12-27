package data

type props struct {
	Protocol protocol `mapstructure:"protocol"`
	Mapping  mapping  `mapstructure:"mapping"`
}

type protocol struct {
	Mode     string   `mapstructure:"mode"`
	Postgres postgres `mapstructure:"postgres"`
}

type mapping struct {
	Modes []string `mapstructure:"modes"`
}

type postgres struct {
	Url string `mapstructure:"url"`
}
