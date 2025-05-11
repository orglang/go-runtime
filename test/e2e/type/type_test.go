package type_test

import (
	"os"
	"testing"

	typedef "smecalculus/rolevod/app/type/def"
)

var (
	api = typedef.NewAPI()
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
