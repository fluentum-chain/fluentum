package version

import (
	"fmt"
	"strconv"
	"strings"
)

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

// parseVersion parses a semantic version string and returns major, minor, patch
func parseVersion(v string) (major, minor, patch int, err error) {
	// Remove 'v' prefix if present
	v = strings.TrimPrefix(v, "v")

	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", v)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return major, minor, patch, nil
}

// Compatible checks if the current version is compatible with the required minimum version
func Compatible(minVersion string) bool {
	currentMajor, currentMinor, currentPatch, err := parseVersion(Version)
	if err != nil {
		// If we can't parse current version, assume incompatible
		return false
	}

	minMajor, minMinor, minPatch, err := parseVersion(minVersion)
	if err != nil {
		// If we can't parse minimum version, assume incompatible
		return false
	}

	// Major version must match exactly
	if currentMajor != minMajor {
		return false
	}

	// Current minor must be >= minimum minor
	if currentMinor < minMinor {
		return false
	}

	// If minor versions match, current patch must be >= minimum patch
	if currentMinor == minMinor && currentPatch < minPatch {
		return false
	}

	return true
}
