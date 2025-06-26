package types

import (
	"context"
	cometbftabci "github.com/cometbft/cometbft/abci/types"
	cometbftcrypto "github.com/cometbft/cometbft/crypto"
	cometbfttypes "github.com/cometbft/cometbft/types"
)

// ABCI Types - Direct aliases to CometBFT types for full compatibility

// Request types
type Request = cometbftabci.Request
type RequestCheckTx = cometbftabci.RequestCheckTx
type RequestFinalizeBlock = cometbftabci.RequestFinalizeBlock
type RequestCommit = cometbftabci.RequestCommit
type RequestInfo = cometbftabci.RequestInfo
type RequestQuery = cometbftabci.RequestQuery
type RequestInitChain = cometbftabci.RequestInitChain
type RequestPrepareProposal = cometbftabci.RequestPrepareProposal
type RequestProcessProposal = cometbftabci.RequestProcessProposal
type RequestExtendVote = cometbftabci.RequestExtendVote
type RequestVerifyVoteExtension = cometbftabci.RequestVerifyVoteExtension
type RequestListSnapshots = cometbftabci.RequestListSnapshots
type RequestOfferSnapshot = cometbftabci.RequestOfferSnapshot
type RequestLoadSnapshotChunk = cometbftabci.RequestLoadSnapshotChunk
type RequestApplySnapshotChunk = cometbftabci.RequestApplySnapshotChunk

// Response types
type Response = cometbftabci.Response
type ResponseCheckTx = cometbftabci.ResponseCheckTx
type ResponseFinalizeBlock = cometbftabci.ResponseFinalizeBlock
type ResponseCommit = cometbftabci.ResponseCommit
type ResponseInfo = cometbftabci.ResponseInfo
type ResponseQuery = cometbftabci.ResponseQuery
type ResponseInitChain = cometbftabci.ResponseInitChain
type ResponsePrepareProposal = cometbftabci.ResponsePrepareProposal
type ResponseProcessProposal = cometbftabci.ResponseProcessProposal
type ResponseExtendVote = cometbftabci.ResponseExtendVote
type ResponseVerifyVoteExtension = cometbftabci.ResponseVerifyVoteExtension
type ResponseListSnapshots = cometbftabci.ResponseListSnapshots
type ResponseOfferSnapshot = cometbftabci.ResponseOfferSnapshot
type ResponseLoadSnapshotChunk = cometbftabci.ResponseLoadSnapshotChunk
type ResponseApplySnapshotChunk = cometbftabci.ResponseApplySnapshotChunk

// Consensus and block types
type ConsensusParams = cometbftabci.ConsensusParams
type BlockParams = cometbftabci.BlockParams
type EvidenceParams = cometbftabci.EvidenceParams
type ValidatorParams = cometbftabci.ValidatorParams
type VersionParams = cometbftabci.VersionParams
type Header = cometbfttypes.Header
type BlockID = cometbfttypes.BlockID
type PartSetHeader = cometbfttypes.PartSetHeader

// Validator types
type Validator = cometbftabci.Validator
type ValidatorUpdate = cometbftabci.ValidatorUpdate
type PubKey = cometbftcrypto.PubKey

// Transaction and execution types
type ExecTxResult = cometbftabci.ExecTxResult
type Event = cometbftabci.Event
type EventAttribute = cometbftabci.EventAttribute
type TxResult = cometbftabci.TxResult

// Snapshot types
type Snapshot = cometbftabci.Snapshot

// Vote and commit types
type VoteInfo = cometbftabci.VoteInfo
type ExtendedVoteInfo = cometbftabci.ExtendedVoteInfo
type CommitInfo = cometbftabci.CommitInfo
type ExtendedCommitInfo = cometbftabci.ExtendedCommitInfo
type Misbehavior = cometbftabci.Misbehavior

// Proof types
type ProofOp = cometbftabci.ProofOp
type ProofOps = cometbftabci.ProofOps

// CheckTx types
type CheckTxType = cometbftabci.CheckTxType

// Response status types
type ResponseProcessProposal_Status = cometbftabci.ResponseProcessProposal_Status
type ResponseVerifyVoteExtension_Status = cometbftabci.ResponseVerifyVoteExtension_Status
type ResponseOfferSnapshot_Result = cometbftabci.ResponseOfferSnapshot_Result
type ResponseApplySnapshotChunk_Result = cometbftabci.ResponseApplySnapshotChunk_Result

// Constants for response statuses
const (
	ResponseProcessProposal_ACCEPT  = cometbftabci.ResponseProcessProposal_ACCEPT
	ResponseProcessProposal_REJECT  = cometbftabci.ResponseProcessProposal_REJECT
	ResponseVerifyVoteExtension_ACCEPT = cometbftabci.ResponseVerifyVoteExtension_ACCEPT
	ResponseVerifyVoteExtension_REJECT = cometbftabci.ResponseVerifyVoteExtension_REJECT
	ResponseOfferSnapshot_ACCEPT   = cometbftabci.ResponseOfferSnapshot_ACCEPT
	ResponseOfferSnapshot_REJECT   = cometbftabci.ResponseOfferSnapshot_REJECT
	ResponseApplySnapshotChunk_ACCEPT = cometbftabci.ResponseApplySnapshotChunk_ACCEPT
	ResponseApplySnapshotChunk_REJECT = cometbftabci.ResponseApplySnapshotChunk_REJECT
)

// CheckTx type constants
const (
	CheckTxType_New     = cometbftabci.CheckTxType_New
	CheckTxType_Recheck = cometbftabci.CheckTxType_Recheck
)

// Error codes (using CometBFT's constants)
const (
	CodeTypeOK                = cometbftabci.CodeTypeOK
	CodeTypeInternalError     = cometbftabci.CodeTypeInternalError
	CodeTypeEncodingError     = cometbftabci.CodeTypeEncodingError
	CodeTypeUnauthorized      = cometbftabci.CodeTypeUnauthorized
	CodeTypeInsufficientFunds = cometbftabci.CodeTypeInsufficientFunds
	CodeTypeUnknownRequest    = cometbftabci.CodeTypeUnknownRequest
	CodeTypeInvalidAddress    = cometbftabci.CodeTypeInvalidAddress
)

// CometBFT core types (for compatibility)
type CometBFTHeader = cometbfttypes.Header
type CometBFTBlockID = cometbfttypes.BlockID
type CometBFTPartSetHeader = cometbfttypes.PartSetHeader

// Helper functions for type conversions

// ToCometBFTHeader converts our Header to CometBFT Header
func ToCometBFTHeader(h Header) cometbfttypes.Header {
	return cometbfttypes.Header{
		Version:            h.Version,
		ChainID:            h.ChainID,
		Height:             h.Height,
		Time:               h.Time,
		LastBlockID:        ToCometBFTBlockID(h.LastBlockID),
		LastCommitHash:     h.LastCommitHash,
		DataHash:           h.DataHash,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		LastResultsHash:    h.LastResultsHash,
		EvidenceHash:       h.EvidenceHash,
		ProposerAddress:    h.ProposerAddress,
	}
}

// FromCometBFTHeader converts CometBFT Header to our Header
func FromCometBFTHeader(h cometbfttypes.Header) Header {
	return Header{
		Version:            h.Version,
		ChainID:            h.ChainID,
		Height:             h.Height,
		Time:               h.Time,
		LastBlockID:        FromCometBFTBlockID(h.LastBlockID),
		LastCommitHash:     h.LastCommitHash,
		DataHash:           h.DataHash,
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		AppHash:            h.AppHash,
		LastResultsHash:    h.LastResultsHash,
		EvidenceHash:       h.EvidenceHash,
		ProposerAddress:    h.ProposerAddress,
	}
}

// ToCometBFTBlockID converts our BlockID to CometBFT BlockID
func ToCometBFTBlockID(b BlockID) cometbfttypes.BlockID {
	return cometbfttypes.BlockID{
		Hash:          b.Hash,
		PartSetHeader: ToCometBFTPartSetHeader(b.PartSetHeader),
	}
}

// FromCometBFTBlockID converts CometBFT BlockID to our BlockID
func FromCometBFTBlockID(b cometbfttypes.BlockID) BlockID {
	return BlockID{
		Hash:          b.Hash,
		PartSetHeader: FromCometBFTPartSetHeader(b.PartSetHeader),
	}
}

// ToCometBFTPartSetHeader converts our PartSetHeader to CometBFT PartSetHeader
func ToCometBFTPartSetHeader(p PartSetHeader) cometbfttypes.PartSetHeader {
	return cometbfttypes.PartSetHeader{
		Total: p.Total,
		Hash:  p.Hash,
	}
}

// FromCometBFTPartSetHeader converts CometBFT PartSetHeader to our PartSetHeader
func FromCometBFTPartSetHeader(p cometbfttypes.PartSetHeader) PartSetHeader {
	return PartSetHeader{
		Total: p.Total,
		Hash:  p.Hash,
	}
}

// Utility functions for common operations

// IsOK checks if a response code indicates success
func IsOK(code uint32) bool {
	return code == CodeTypeOK
}

// IsError checks if a response code indicates an error
func IsError(code uint32) bool {
	return code != CodeTypeOK
}

// NewEvent creates a new event with the given type
func NewEvent(eventType string) Event {
	return Event{
		Type: eventType,
	}
}

// AddAttribute adds an attribute to an event
func AddAttribute(event *Event, key, value string, index bool) {
	event.Attributes = append(event.Attributes, EventAttribute{
		Key:   key,
		Value: value,
		Index: index,
	})
}

// NewExecTxResult creates a new ExecTxResult
func NewExecTxResult(code uint32, data []byte, log string) *ExecTxResult {
	return &ExecTxResult{
		Code: code,
		Data: data,
		Log:  log,
	}
}

// NewResponseCheckTx creates a new ResponseCheckTx
func NewResponseCheckTx(code uint32, data []byte, log string, gasWanted, gasUsed int64) *ResponseCheckTx {
	return &ResponseCheckTx{
		Code:      code,
		Data:      data,
		Log:       log,
		GasWanted: gasWanted,
		GasUsed:   gasUsed,
	}
}

// NewResponseQuery creates a new ResponseQuery
func NewResponseQuery(code uint32, value []byte, log string) *ResponseQuery {
	return &ResponseQuery{
		Code:  code,
		Value: value,
		Log:   log,
	}
}

// Application interface defines the ABCI application methods
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