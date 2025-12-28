package dec

import (
	"github.com/go-resty/resty/v2"
)

// Client-side secondary adapter
type sdkResty struct {
	resty *resty.Client
}

func newSdkResty() *sdkResty {
	r := resty.New().SetBaseURL("http://localhost:8080/api/v1")
	return &sdkResty{r}
}

func NewAPI() API {
	return newSdkResty()
}

func (cl *sdkResty) Create(spec PoolSpec) (PoolRef, error) {
	return PoolRef{}, nil
}
