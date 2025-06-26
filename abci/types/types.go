package types

import (
	"context"
	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// Type aliases for all ABCI types - using CometBFT types directly
type (
	// Core types
	ValidatorUpdate       = cmtabci.ValidatorUpdate
	Event                 = cmtabci.Event
	EventAttribute        = cmtabci.EventAttribute
	ExecTxResult          = cmtabci.ExecTxResult
	RequestCheckTx        = cmtabci.RequestCheckTx
	ResponseCheckTx       = cmtabci.ResponseCheckTx
	RequestFinalizeBlock  = cmtabci.RequestFinalizeBlock
	ResponseFinalizeBlock = cmtabci.ResponseFinalizeBlock
	RequestCommit         = cmtabci.RequestCommit
	ResponseCommit        = cmtabci.ResponseCommit
	RequestInfo           = cmtabci.RequestInfo
	ResponseInfo          = cmtabci.ResponseInfo
	RequestQuery          = cmtabci.RequestQuery
	ResponseQuery         = cmtabci.ResponseQuery
	RequestInitChain      = cmtabci.RequestInitChain
	ResponseInitChain     = cmtabci.ResponseInitChain
	
	// Proposal types
	RequestPrepareProposal  = cmtabci.RequestPrepareProposal
	ResponsePrepareProposal = cmtabci.ResponsePrepareProposal
	RequestProcessProposal  = cmtabci.RequestProcessProposal
	ResponseProcessProposal = cmtabci.ResponseProcessProposal
	
	// Vote extension types
	RequestExtendVote             = cmtabci.RequestExtendVote
	ResponseExtendVote            = cmtabci.ResponseExtendVote
	RequestVerifyVoteExtension    = cmtabci.RequestVerifyVoteExtension
	ResponseVerifyVoteExtension   = cmtabci.ResponseVerifyVoteExtension
	
	// Snapshot types
	RequestListSnapshots      = cmtabci.RequestListSnapshots
	ResponseListSnapshots     = cmtabci.ResponseListSnapshots
	RequestOfferSnapshot      = cmtabci.RequestOfferSnapshot
	ResponseOfferSnapshot     = cmtabci.ResponseOfferSnapshot
	RequestLoadSnapshotChunk  = cmtabci.RequestLoadSnapshotChunk
	ResponseLoadSnapshotChunk = cmtabci.ResponseLoadSnapshotChunk
	RequestApplySnapshotChunk = cmtabci.RequestApplySnapshotChunk
	ResponseApplySnapshotChunk = cmtabci.ResponseApplySnapshotChunk
	Snapshot                  = cmtabci.Snapshot
	
	// Consensus types
	ConsensusParams = cmtabci.ConsensusParams
	Header          = cmtabci.Header
	BlockID         = cmtabci.BlockID
	PartSetHeader   = cmtabci.PartSetHeader
	
	// Enums
	CheckTxType = cmtabci.CheckTxType
)

// Response codes - using CometBFT constants
const (
	CodeTypeOK                = cmtabci.CodeTypeOK
	CodeTypeInternalError     = cmtabci.CodeTypeInternalError
	CodeTypeEncodingError     = cmtabci.CodeTypeEncodingError
	CodeTypeUnauthorized      = cmtabci.CodeTypeUnauthorized
	CodeTypeInsufficientFunds = cmtabci.CodeTypeInsufficientFunds
	CodeTypeUnknownRequest    = cmtabci.CodeTypeUnknownRequest
	CodeTypeInvalidAddress    = cmtabci.CodeTypeInvalidAddress
	CodeTypeInvalidPubKey     = cmtabci.CodeTypeInvalidPubKey
	CodeTypeUnknownAddress    = cmtabci.CodeTypeUnknownAddress
	CodeTypeInsufficientCoins = cmtabci.CodeTypeInsufficientCoins
	CodeTypeInvalidCoins      = cmtabci.CodeTypeInvalidCoins
	CodeTypeOutOfGas          = cmtabci.CodeTypeOutOfGas
	CodeTypeMemoTooLarge      = cmtabci.CodeTypeMemoTooLarge
	CodeTypeInsufficientFee   = cmtabci.CodeTypeInsufficientFee
	CodeTypeTooManySignatures = cmtabci.CodeTypeTooManySignatures
	CodeTypeNoSignatures      = cmtabci.CodeTypeNoSignatures
	CodeTypeErr               = cmtabci.CodeTypeErr
)

// CheckTxType constants
const (
	CheckTxType_New     = cmtabci.CheckTxType_New
	CheckTxType_Recheck = cmtabci.CheckTxType_Recheck
)

// Application interface matches CometBFT's expected interface exactly
type Application interface {
	// Info/Query Connection
	Info(context.Context, *RequestInfo) (*ResponseInfo, error)
	Query(context.Context, *RequestQuery) (*ResponseQuery, error)

	// Mempool Connection
	CheckTx(context.Context, *RequestCheckTx) (*ResponseCheckTx, error)

	// Consensus Connection
	PrepareProposal(context.Context, *RequestPrepareProposal) (*ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *RequestProcessProposal) (*ResponseProcessProposal, error)
	FinalizeBlock(context.Context, *RequestFinalizeBlock) (*ResponseFinalizeBlock, error)
	ExtendVote(context.Context, *RequestExtendVote) (*ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *RequestVerifyVoteExtension) (*ResponseVerifyVoteExtension, error)
	Commit(context.Context, *RequestCommit) (*ResponseCommit, error)
	InitChain(context.Context, *RequestInitChain) (*ResponseInitChain, error)

	// State Sync Connection
	ListSnapshots(context.Context, *RequestListSnapshots) (*ResponseListSnapshots, error)
	OfferSnapshot(context.Context, *RequestOfferSnapshot) (*ResponseOfferSnapshot, error)
	LoadSnapshotChunk(context.Context, *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error)
	ApplySnapshotChunk(context.Context, *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error)
}

// Helper functions
func IsOK(code uint32) bool {
	return cmtabci.IsOK(code)
}

func IsError(code uint32) bool {
	return cmtabci.IsError(code)
} 