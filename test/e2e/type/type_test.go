package type_test

import (
	"os"
	"testing"

	typedef "orglang/orglang/aat/type/def"
)

var (
	api = typedef.NewAPI()
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
