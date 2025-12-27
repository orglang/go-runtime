package sig_test

import (
	"os"
	"testing"

	procdec "orglang/orglang/aat/proc/dec"
)

var (
	api = procdec.NewAPI()
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
