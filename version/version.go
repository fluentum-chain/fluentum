package version

var TMCoreSemVer = TMVersionDefault

const (
	// TMVersionDefault is the used as the fallback version of Tendermint Core
	// when not using git describe. It is formatted with semantic versioning.
	TMVersionDefault = "0.34.24"
	// ABCISemVer is the semantic version of the ABCI library
	ABCISemVer = "0.17.0"

	ABCIVersion = ABCISemVer
)

var (
	// P2PProtocol versions all p2p behaviour and msgs.
	// This includes proposer selection.
	P2PProtocol uint64 = 8

	// BlockProtocol versions all block data structures and processing.
	// This includes validity of blocks and state updates.
	BlockProtocol uint64 = 11
)

// Version is the current version of Fluentum Core
var Version = "v0.1.0"

// GitCommit is the git commit hash
var GitCommit = ""

// BuildTime is the build timestamp
var BuildTime = ""

// GoVersion is the Go version used to build
var GoVersion = ""

// BuildTags are the build tags used
var BuildTags = ""
