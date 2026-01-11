package rc

import (
	"github.com/go-resty/resty/v2"
)

func newRestyClient() *resty.Client {
	return resty.New().SetBaseURL("http://localhost:8080/api/v1")
}
