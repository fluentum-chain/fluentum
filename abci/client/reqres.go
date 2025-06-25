package client

import (
	"github.com/cometbft/cometbft/abci/types"
)

type ReqRes struct {
	Request    *types.Request
	Response   *types.Response
	Error      error
	ResponseCb func(*types.Response, error)
	DoneCh     chan struct{}
	ResponseCh chan interface{}
	ErrorCh    chan error
}

func NewReqRes(req *types.Request) *ReqRes {
	return &ReqRes{
		Request:    req,
		DoneCh:     make(chan struct{}),
		ResponseCh: make(chan interface{}, 1),
		ErrorCh:    make(chan error, 1),
	}
}
