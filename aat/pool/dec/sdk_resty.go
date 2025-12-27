package dec

import (
	"github.com/go-resty/resty/v2"
)

// Adapter
type clientResty struct {
	resty *resty.Client
}

func newClientResty() *clientResty {
	r := resty.New().SetBaseURL("http://localhost:8080/api/v1")
	return &clientResty{r}
}

func NewAPI() API {
	return newClientResty()
}

func (cl *clientResty) Create(spec PoolSpec) (PoolRef, error) {
	return PoolRef{}, nil
}
