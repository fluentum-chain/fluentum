package p2p

import (
	"github.com/gogo/protobuf/proto"
)

// Wrap implements the Wrapper interface for PexRequest
func (m *PexRequest) Wrap() proto.Message {
	pm := &Message{}
	pm.Sum = &Message_PexRequest{PexRequest: m}
	return pm
}

// Wrap implements the Wrapper interface for PexAddrs
func (m *PexAddrs) Wrap() proto.Message {
	pm := &Message{}
	pm.Sum = &Message_PexAddrs{PexAddrs: m}
	return pm
}
