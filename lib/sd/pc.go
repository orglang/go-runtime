package sd

import (
	"orglang/orglang/lib/cs"
)

func newStoragePC(k cs.Keeper) (storagePC, error) {
	pc := &storagePC{}
	loadingErr := k.Load("storage", pc)
	if loadingErr != nil {
		return storagePC{}, loadingErr
	}
	validationErr := pc.Validate()
	if validationErr != nil {
		return storagePC{}, validationErr
	}
	return *pc, nil
}

type storagePC struct {
	Protocol protocolPC `mapstructure:"protocol"`
	Driver   driverPC   `mapstructure:"driver"`
}

type protocolPC struct {
	Mode     protocolMode `mapstructure:"mode"`
	Postgres postgresPC   `mapstructure:"postgres"`
}

type driverPC struct {
	Mode driverMode `mapstructure:"mode"`
	Pgx  pgxPC      `mapstructure:"pgx"`
}

type postgresPC struct {
	Url string `mapstructure:"url"`
}

type pgxPC struct{}

type protocolMode string

const (
	postgresMode = protocolMode("postgres")
)

type driverMode string

const (
	pgxMode = driverMode("pgx")
)
