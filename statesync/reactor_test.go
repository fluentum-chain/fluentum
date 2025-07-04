package statesync

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/p2p"
	p2pmocks "github.com/fluentum-chain/fluentum/p2p/mocks"
	ssproto "github.com/fluentum-chain/fluentum/proto/tendermint/statesync"
	proxymocks "github.com/fluentum-chain/fluentum/proxy/mocks"
)

func TestReactor_Receive_ChunkRequest(t *testing.T) {
	testcases := map[string]struct {
		request        *ssproto.ChunkRequest
		chunk          []byte
		expectResponse *ssproto.ChunkResponse
	}{
		"chunk is returned": {
			&ssproto.ChunkRequest{Height: 1, Format: 1, Index: 1},
			[]byte{1, 2, 3},
			&ssproto.ChunkResponse{Height: 1, Format: 1, Index: 1, Chunk: []byte{1, 2, 3}}},
		"empty chunk is returned, as nil": {
			&ssproto.ChunkRequest{Height: 1, Format: 1, Index: 1},
			[]byte{},
			&ssproto.ChunkResponse{Height: 1, Format: 1, Index: 1, Chunk: nil}},
		"nil (missing) chunk is returned as missing": {
			&ssproto.ChunkRequest{Height: 1, Format: 1, Index: 1},
			nil,
			&ssproto.ChunkResponse{Height: 1, Format: 1, Index: 1, Missing: true},
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// Mock ABCI connection to return local snapshots
			conn := &proxymocks.AppConnSnapshot{}
			conn.On("LoadSnapshotChunkSync", abci.RequestLoadSnapshotChunk{
				Height: tc.request.Height,
				Format: tc.request.Format,
				Chunk:  tc.request.Index,
			}).Return(&abci.ResponseLoadSnapshotChunk{Chunk: tc.chunk}, nil)

			// Mock peer to store response, if found
			peer := &p2pmocks.Peer{}
			peer.On("ID").Return(p2p.ID("id"))
			var response *ssproto.ChunkResponse
			if tc.expectResponse != nil {
				peer.On("SendEnvelope", mock.MatchedBy(func(i interface{}) bool {
					e, ok := i.(p2p.Envelope)
					return ok && e.ChannelID == ChunkChannel
				})).Run(func(args mock.Arguments) {
					e := args[0].(p2p.Envelope)

					// Marshal to simulate a wire roundtrip.
					bz, err := proto.Marshal(e.Message)
					require.NoError(t, err)
					err = proto.Unmarshal(bz, e.Message)
					require.NoError(t, err)
					response = e.Message.(*ssproto.ChunkResponse)
				}).Return(true)
			}

			// Start a reactor and send a ssproto.ChunkRequest, then wait for and check response
			cfg := config.DefaultStateSyncConfig()
			r := NewReactor(*cfg, conn, nil, "")
			err := r.Start()
			require.NoError(t, err)
			t.Cleanup(func() {
				if err := r.Stop(); err != nil {
					t.Error(err)
				}
			})

			r.ReceiveEnvelope(p2p.Envelope{
				ChannelID: ChunkChannel,
				Src:       peer,
				Message:   tc.request,
			})
			time.Sleep(100 * time.Millisecond)
			assert.Equal(t, tc.expectResponse, response)

			conn.AssertExpectations(t)
			peer.AssertExpectations(t)
		})
	}
}

func TestReactor_Receive_SnapshotsRequest(t *testing.T) {
	testcases := map[string]struct {
		snapshots       []*abci.Snapshot
		expectResponses []*ssproto.SnapshotsResponse
	}{
		"no snapshots": {nil, []*ssproto.SnapshotsResponse{}},
		">10 unordered snapshots": {
			[]*abci.Snapshot{
				{Height: 1, Format: 2, Chunks: 7, Hash: []byte{1, 2}, Metadata: []byte{1}},
				{Height: 2, Format: 2, Chunks: 7, Hash: []byte{2, 2}, Metadata: []byte{2}},
				{Height: 3, Format: 2, Chunks: 7, Hash: []byte{3, 2}, Metadata: []byte{3}},
				{Height: 1, Format: 1, Chunks: 7, Hash: []byte{1, 1}, Metadata: []byte{4}},
				{Height: 2, Format: 1, Chunks: 7, Hash: []byte{2, 1}, Metadata: []byte{5}},
				{Height: 3, Format: 1, Chunks: 7, Hash: []byte{3, 1}, Metadata: []byte{6}},
				{Height: 1, Format: 4, Chunks: 7, Hash: []byte{1, 4}, Metadata: []byte{7}},
				{Height: 2, Format: 4, Chunks: 7, Hash: []byte{2, 4}, Metadata: []byte{8}},
				{Height: 3, Format: 4, Chunks: 7, Hash: []byte{3, 4}, Metadata: []byte{9}},
				{Height: 1, Format: 3, Chunks: 7, Hash: []byte{1, 3}, Metadata: []byte{10}},
				{Height: 2, Format: 3, Chunks: 7, Hash: []byte{2, 3}, Metadata: []byte{11}},
				{Height: 3, Format: 3, Chunks: 7, Hash: []byte{3, 3}, Metadata: []byte{12}},
			},
			[]*ssproto.SnapshotsResponse{
				{Height: 3, Format: 4, Chunks: 7, Hash: []byte{3, 4}, Metadata: []byte{9}},
				{Height: 3, Format: 3, Chunks: 7, Hash: []byte{3, 3}, Metadata: []byte{12}},
				{Height: 3, Format: 2, Chunks: 7, Hash: []byte{3, 2}, Metadata: []byte{3}},
				{Height: 3, Format: 1, Chunks: 7, Hash: []byte{3, 1}, Metadata: []byte{6}},
				{Height: 2, Format: 4, Chunks: 7, Hash: []byte{2, 4}, Metadata: []byte{8}},
				{Height: 2, Format: 3, Chunks: 7, Hash: []byte{2, 3}, Metadata: []byte{11}},
				{Height: 2, Format: 2, Chunks: 7, Hash: []byte{2, 2}, Metadata: []byte{2}},
				{Height: 2, Format: 1, Chunks: 7, Hash: []byte{2, 1}, Metadata: []byte{5}},
				{Height: 1, Format: 4, Chunks: 7, Hash: []byte{1, 4}, Metadata: []byte{7}},
				{Height: 1, Format: 3, Chunks: 7, Hash: []byte{1, 3}, Metadata: []byte{10}},
			},
		},
	}

	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// Mock ABCI connection to return local snapshots
			conn := &proxymocks.AppConnSnapshot{}
			conn.On("ListSnapshotsSync", abci.RequestListSnapshots{}).Return(&abci.ResponseListSnapshots{
				Snapshots: tc.snapshots,
			}, nil)

			// Mock peer to catch responses and store them in a slice
			responses := []*ssproto.SnapshotsResponse{}
			peer := &p2pmocks.Peer{}
			if len(tc.expectResponses) > 0 {
				peer.On("ID").Return(p2p.ID("id"))
				peer.On("SendEnvelope", mock.MatchedBy(func(i interface{}) bool {
					e, ok := i.(p2p.Envelope)
					return ok && e.ChannelID == SnapshotChannel
				})).Run(func(args mock.Arguments) {
					e := args[0].(p2p.Envelope)

					// Marshal to simulate a wire roundtrip.
					bz, err := proto.Marshal(e.Message)
					require.NoError(t, err)
					err = proto.Unmarshal(bz, e.Message)
					require.NoError(t, err)
					responses = append(responses, e.Message.(*ssproto.SnapshotsResponse))
				}).Return(true)
			}

			// Start a reactor and send a SnapshotsRequestMessage, then wait for and check responses
			cfg := config.DefaultStateSyncConfig()
			r := NewReactor(*cfg, conn, nil, "")
			err := r.Start()
			require.NoError(t, err)
			t.Cleanup(func() {
				if err := r.Stop(); err != nil {
					t.Error(err)
				}
			})

			r.ReceiveEnvelope(p2p.Envelope{
				ChannelID: SnapshotChannel,
				Src:       peer,
				Message:   &ssproto.SnapshotsRequest{},
			})
			time.Sleep(100 * time.Millisecond)
			assert.Equal(t, tc.expectResponses, responses)

			conn.AssertExpectations(t)
			peer.AssertExpectations(t)
		})
	}
}

func TestLegacyReactorReceiveBasic(t *testing.T) {
	cfg := config.DefaultStateSyncConfig()
	conn := &proxymocks.AppConnSnapshot{}
	reactor := NewReactor(*cfg, conn, nil, "")
	peer := p2p.CreateRandomPeer(false)

	reactor.InitPeer(peer)
	reactor.AddPeer(peer)
	m := &ssproto.ChunkRequest{Height: 1, Format: 1, Index: 1}
	wm := m.Wrap()
	msg, err := proto.Marshal(wm)
	assert.NoError(t, err)

	assert.NotPanics(t, func() {
		reactor.Receive(ChunkChannel, peer, msg)
	})
}
