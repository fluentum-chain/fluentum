package types

import (
	"time"
)

// FinalizeBlock replaces BeginBlock/DeliverTx/EndBlock
type RequestFinalizeBlock struct {
	Height int64
	Txs    [][]byte
	Hash   []byte
	Header *Header
}

type ResponseFinalizeBlock struct {
	TxResults             []*ExecTxResult
	ValidatorUpdates      []ValidatorUpdate
	ConsensusParamUpdates *ConsensusParams
	AppHash               []byte
	Events                []Event
}

type ExecTxResult struct {
	Code      uint32
	Data      []byte
	Log       string
	Info      string
	Events    []Event
	GasUsed   int64
	GasWanted int64
}

// Proposal handling
type RequestPrepareProposal struct {
	MaxTxBytes int64
	Txs        [][]byte
}

type ResponsePrepareProposal struct {
	Txs [][]byte
}

type RequestProcessProposal struct {
	Height int64
	Txs    [][]byte
	Hash   []byte
	Header *Header
}

type ResponseProcessProposal struct {
	Status ResponseProcessProposal_Status
}

type ResponseProcessProposal_Status int32

const (
	ResponseProcessProposal_UNKNOWN ResponseProcessProposal_Status = 0
	ResponseProcessProposal_ACCEPT  ResponseProcessProposal_Status = 1
	ResponseProcessProposal_REJECT  ResponseProcessProposal_Status = 2
)

// Vote extension types
type RequestExtendVote struct {
	Height int64
	Round  int32
	Hash   []byte
}

type ResponseExtendVote struct {
	VoteExtension []byte
}

type RequestVerifyVoteExtension struct {
	Height        int64
	Round         int32
	Hash          []byte
	VoteExtension []byte
}

type ResponseVerifyVoteExtension struct {
	Status ResponseVerifyVoteExtension_Status
}

type ResponseVerifyVoteExtension_Status int32

const (
	ResponseVerifyVoteExtension_UNKNOWN ResponseVerifyVoteExtension_Status = 0
	ResponseVerifyVoteExtension_ACCEPT  ResponseVerifyVoteExtension_Status = 1
	ResponseVerifyVoteExtension_REJECT  ResponseVerifyVoteExtension_Status = 2
)

// Commit types
type RequestCommit struct{}

type ResponseCommit struct {
	Data         []byte
	RetainHeight int64
}

// InitChain types
type RequestInitChain struct {
	Time            time.Time
	ChainId         string
	ConsensusParams *ConsensusParams
	Validators      []ValidatorUpdate
	AppStateBytes   []byte
	InitialHeight   int64
}

type ResponseInitChain struct {
	ConsensusParams *ConsensusParams
	Validators      []ValidatorUpdate
	AppHash         []byte
}

// Header type for block information
type Header struct {
	Version            Version
	ChainID            string
	Height             int64
	Time               time.Time
	LastBlockId        BlockID
	LastCommitHash     []byte
	DataHash           []byte
	ValidatorsHash     []byte
	NextValidatorsHash []byte
	ConsensusHash      []byte
	AppHash            []byte
	LastResultsHash    []byte
	EvidenceHash       []byte
	ProposerAddress    []byte
}

type Version struct {
	Block uint64
	App   uint64
}

type BlockID struct {
	Hash          []byte
	PartSetHeader PartSetHeader
}

type PartSetHeader struct {
	Total uint32
	Hash  []byte
}

// ConsensusParams type
type ConsensusParams struct {
	Block     *BlockParams
	Evidence  *EvidenceParams
	Validator *ValidatorParams
	Version   *VersionParams
}

type BlockParams struct {
	MaxBytes int64
	MaxGas   int64
}

type EvidenceParams struct {
	MaxAgeNumBlocks int64
	MaxAgeDuration  time.Duration
	MaxBytes        int64
}

type ValidatorParams struct {
	PubKeyTypes []string
}

type VersionParams struct {
	AppVersion uint64
} 