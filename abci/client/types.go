package client

import (
	"context"
	cmtabci "github.com/cometbft/cometbft/abci/types"
)

type (
	ReqRes struct {
		Request    *cmtabci.Request
		Response   *cmtabci.Response
		Error      error
		ResponseCb func(*cmtabci.Response, error)
		DoneCh     chan struct{}
		ResponseCh chan interface{}
		ErrorCh    chan error
	}

	Callback func(*cmtabci.Request, *cmtabci.Response)

	Logger interface {
		Debug(msg string, keyVals ...interface{})
		Info(msg string, keyVals ...interface{})
		Error(msg string, keyVals ...interface{})
	}
)

func NewReqRes(req *cmtabci.Request) *ReqRes {
	return &ReqRes{
		Request:    req,
		DoneCh:     make(chan struct{}),
		ResponseCh: make(chan interface{}, 1),
		ErrorCh:    make(chan error, 1),
	}
}

// Done marks the request as done
func (reqRes *ReqRes) Done() {
	close(reqRes.DoneCh)
}

// InvokeCallback invokes the response callback if set
func (reqRes *ReqRes) InvokeCallback() {
	if reqRes.ResponseCb != nil {
		reqRes.ResponseCb(reqRes.Response, reqRes.Error)
	}
}

// Wait waits for the request to complete
func (reqRes *ReqRes) Wait() {
	<-reqRes.DoneCh
} 
