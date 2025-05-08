package type_test

import (
	"os"
	"testing"

	typedec "smecalculus/rolevod/app/type/dec"
)

var (
	api = typedec.NewAPI()
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
