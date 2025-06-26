package types

// Info types
type RequestInfo struct {
	Version      string
	BlockVersion uint64
	P2PVersion   uint64
	AbciVersion  uint64
}

type ResponseInfo struct {
	Data             string
	Version          string
	AppVersion       uint64
	LastBlockHeight  int64
	LastBlockAppHash []byte
}

// Query types
type RequestQuery struct {
	Data   []byte
	Path   string
	Height int64
	Prove  bool
}

type ResponseQuery struct {
	Code      uint32
	Log       string
	Info      string
	Index     int64
	Key       []byte
	Value     []byte
	ProofOps  *ProofOps
	Height    int64
	Codespace string
}

// ProofOps type for query proofs
type ProofOps struct {
	Ops []ProofOp
}

type ProofOp struct {
	Type string
	Key  []byte
	Data []byte
} 