package procdec

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

func (sdk *RestySDK) Incept(decQN qualsym.ADT) (DecRef, error) {
	return DecRef{}, nil
}

func (sdk *RestySDK) Create(spec DecSpec) (DecSnap, error) {
	req := MsgFromDecSpec(spec)
	var res DecSnapME
	resp, err := sdk.Client.R().
		SetResult(&res).
		SetBody(&req).
		Post("/declarations")
	if err != nil {
		return DecSnap{}, err
	}
	if resp.IsError() {
		return DecSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToDecSnap(res)
}

func (sdk *RestySDK) RetrieveSnap(id identity.ADT) (DecSnap, error) {
	var res DecSnapME
	resp, err := sdk.Client.R().
		SetResult(&res).
		SetPathParam("id", id.String()).
		Get("/declarations/{id}")
	if err != nil {
		return DecSnap{}, err
	}
	if resp.IsError() {
		return DecSnap{}, fmt.Errorf("received: %v", string(resp.Body()))
	}
	return MsgToDecSnap(res)
}

func (sdk *RestySDK) RetreiveRefs() ([]DecRef, error) {
	refs := []DecRef{}
	return refs, nil
}
