package types

// RequestCheckTx contains the transaction and metadata
type RequestCheckTx struct {
	Tx   []byte
	Type CheckTxType // NEW=0, RECHECK=1
}

type ResponseCheckTx struct {
	Code      uint32
	Data      []byte
	Log       string
	Info      string
	GasWanted int64
	GasUsed   int64
	Events    []Event
	Codespace string
}

type CheckTxType uint32

const (
	CheckTxType_New     CheckTxType = 0
	CheckTxType_Recheck CheckTxType = 1
)

// String returns the string representation of CheckTxType
func (c CheckTxType) String() string {
	switch c {
	case CheckTxType_New:
		return "NEW"
	case CheckTxType_Recheck:
		return "RECHECK"
	default:
		return "UNKNOWN"
	}
} 