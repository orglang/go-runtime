package typedef

import (
	"fmt"

	"github.com/go-resty/resty/v2"

	"orglang/orglang/adt/identity"
	"orglang/orglang/adt/qualsym"
)

// Client-side secondary adapter
type RestySDK struct {
	Client *resty.Client
}

func (sdk *RestySDK) Incept(typeQN qualsym.ADT) (DefRef, error) {
	return DefRef{}, nil
}

func (sdk *RestySDK) Create(spec DefSpec) (DefSnap, error) {
	req := MsgFromDefSpec(spec)
	var res DefSnapME
	resp, err := sdk.Client.R().
		SetResult(&res).
		SetBody(&req).
		Post("/types")
	if err != nil {
		return DefSnap{}, err
	}
	if resp.IsError() {
		return DefSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToDefSnap(res)
}

func (sdk *RestySDK) Modify(snap DefSnap) (DefSnap, error) {
	return DefSnap{}, nil
}

func (sdk *RestySDK) Retrieve(defID identity.ADT) (DefSnap, error) {
	return DefSnap{}, nil
}

func (sdk *RestySDK) retrieveSnap(rec DefRec) (DefSnap, error) {
	return DefSnap{}, nil
}

func (sdk *RestySDK) RetreiveRefs() ([]DefRef, error) {
	return []DefRef{}, nil
}
