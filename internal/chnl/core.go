package chnl

import (
	"fmt"

	"smecalculus/rolevod/lib/id"
	"smecalculus/rolevod/lib/sym"
)

func ErrDoesNotExist(want id.ADT) error {
	return fmt.Errorf("channel doesn't exist: %v", want)
}

func ErrMissingInCfg(want sym.ADT) error {
	return fmt.Errorf("channel missing in cfg: %v", want)
}

func ErrMissingInCtx(want sym.ADT) error {
	return fmt.Errorf("channel missing in ctx: %v", want)
}

func ErrAlreadyClosed(got id.ADT) error {
	return fmt.Errorf("channel already closed: %v", got)
}

func ErrNotAnID(got sym.ADT) error {
	return fmt.Errorf("not a channel id: %v", got)
}
