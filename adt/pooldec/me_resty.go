package pooldec

import (
	"github.com/go-resty/resty/v2"
)

// Client-side secondary adapter
type RestySDK struct {
	Client *resty.Client
}

func (sdk *RestySDK) Create(spec DecSpecME) (DecRefME, error) {
	return DecRefME{}, nil
}
