package types

// Snapshot types
type RequestListSnapshots struct{} // No fields in v0.38

type ResponseListSnapshots struct {
	Snapshots []*Snapshot
}

type Snapshot struct {
	Height   uint64
	Format   uint32
	Chunks   uint32
	Hash     []byte
	Metadata []byte
}

type RequestOfferSnapshot struct {
	Snapshot *Snapshot
	AppHash  []byte
}

type ResponseOfferSnapshot struct {
	Result ResponseOfferSnapshot_Result // ACCEPT=0, ABORT=1, etc
}

type ResponseOfferSnapshot_Result int32

const (
	ResponseOfferSnapshot_UNKNOWN ResponseOfferSnapshot_Result = 0
	ResponseOfferSnapshot_ACCEPT  ResponseOfferSnapshot_Result = 1
	ResponseOfferSnapshot_ABORT   ResponseOfferSnapshot_Result = 2
	ResponseOfferSnapshot_REJECT  ResponseOfferSnapshot_Result = 3
	ResponseOfferSnapshot_REJECT_FORMAT ResponseOfferSnapshot_Result = 4
	ResponseOfferSnapshot_REJECT_SENDER ResponseOfferSnapshot_Result = 5
)

type RequestLoadSnapshotChunk struct {
	Height uint64
	Format uint32
	Chunk  uint32
}

type ResponseLoadSnapshotChunk struct {
	Chunk []byte
}

type RequestApplySnapshotChunk struct {
	Index  uint32
	Chunk  []byte
	Sender string
}

type ResponseApplySnapshotChunk struct {
	Result        ResponseApplySnapshotChunk_Result
	RefetchChunks []uint32
	RejectSenders []string
}

type ResponseApplySnapshotChunk_Result int32

const (
	ResponseApplySnapshotChunk_UNKNOWN ResponseApplySnapshotChunk_Result = 0
	ResponseApplySnapshotChunk_ACCEPT  ResponseApplySnapshotChunk_Result = 1
	ResponseApplySnapshotChunk_ABORT   ResponseApplySnapshotChunk_Result = 2
	ResponseApplySnapshotChunk_RETRY   ResponseApplySnapshotChunk_Result = 3
	ResponseApplySnapshotChunk_RETRY_SNAPSHOT ResponseApplySnapshotChunk_Result = 4
	ResponseApplySnapshotChunk_REJECT_SENDER  ResponseApplySnapshotChunk_Result = 5
) 