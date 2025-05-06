package role_test

import (
	"os"
	"testing"

	rolersig "smecalculus/rolevod/app/role/sig"
)

var (
	api = rolersig.NewAPI()
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
