package dec

import (
	"orglang/orglang/lib/sd"
)

// Port
type repo interface {
	Insert(sd.Source, poolRec) error
}
