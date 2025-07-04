package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"

	"context"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/fluentum-chain/fluentum/libs/log"
	tmnet "github.com/fluentum-chain/fluentum/libs/net"
	"github.com/fluentum-chain/fluentum/libs/service"
	tmsync "github.com/fluentum-chain/fluentum/libs/sync"
)

// var maxNumberConnections = 2

type SocketServer struct {
	service.BaseService
	isLoggerSet bool

	proto    string
	addr     string
	listener net.Listener

	connsMtx   tmsync.Mutex
	conns      map[int]net.Conn
	nextConnID int

	appMtx tmsync.Mutex
	app    cmtabci.Application
}

func NewSocketServer(protoAddr string, app cmtabci.Application) service.Service {
	proto, addr := tmnet.ProtocolAndAddress(protoAddr)
	s := &SocketServer{
		proto:    proto,
		addr:     addr,
		listener: nil,
		app:      app,
		conns:    make(map[int]net.Conn),
	}
	s.BaseService = *service.NewBaseService(nil, "ABCIServer", s)
	return s
}

func (s *SocketServer) SetLogger(l tmlog.Logger) {
	s.BaseService.SetLogger(l)
	s.isLoggerSet = true
}

func (s *SocketServer) OnStart() error {
	ln, err := net.Listen(s.proto, s.addr)
	if err != nil {
		return err
	}

	s.listener = ln
	go s.acceptConnectionsRoutine()

	return nil
}

func (s *SocketServer) OnStop() {
	if err := s.listener.Close(); err != nil {
		s.Logger.Error("Error closing listener", "err", err)
	}

	s.connsMtx.Lock()
	defer s.connsMtx.Unlock()
	for id, conn := range s.conns {
		delete(s.conns, id)
		if err := conn.Close(); err != nil {
			s.Logger.Error("Error closing connection", "id", id, "conn", conn, "err", err)
		}
	}
}

func (s *SocketServer) addConn(conn net.Conn) int {
	s.connsMtx.Lock()
	defer s.connsMtx.Unlock()

	connID := s.nextConnID
	s.nextConnID++
	s.conns[connID] = conn

	return connID
}

// deletes conn even if close errs
func (s *SocketServer) rmConn(connID int) error {
	s.connsMtx.Lock()
	defer s.connsMtx.Unlock()

	conn, ok := s.conns[connID]
	if !ok {
		return fmt.Errorf("connection %d does not exist", connID)
	}

	delete(s.conns, connID)
	return conn.Close()
}

func (s *SocketServer) acceptConnectionsRoutine() {
	for {
		// Accept a connection
		s.Logger.Info("Waiting for new connection...")
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.IsRunning() {
				return // Ignore error from listener closing.
			}
			s.Logger.Error("Failed to accept connection", "err", err)
			continue
		}

		s.Logger.Info("Accepted a new connection")

		connID := s.addConn(conn)

		closeConn := make(chan error, 2)                // Push to signal connection closed
		responses := make(chan *cmtabci.Response, 1000) // A channel to buffer responses

		// Read requests from conn and deal with them
		go s.handleRequests(closeConn, conn, responses)
		// Pull responses from 'responses' and write them to conn.
		go s.handleResponses(closeConn, conn, responses)

		// Wait until signal to close connection
		go s.waitForClose(closeConn, connID)
	}
}

func (s *SocketServer) waitForClose(closeConn chan error, connID int) {
	err := <-closeConn
	switch {
	case err == io.EOF:
		s.Logger.Error("Connection was closed by client")
	case err != nil:
		s.Logger.Error("Connection error", "err", err)
	default:
		// never happens
		s.Logger.Error("Connection was closed")
	}

	// Close the connection
	if err := s.rmConn(connID); err != nil {
		s.Logger.Error("Error closing connection", "err", err)
	}
}

// Read requests from conn and deal with them
func (s *SocketServer) handleRequests(closeConn chan error, conn io.Reader, responses chan<- *cmtabci.Response) {
	var count int
	var bufReader = bufio.NewReader(conn)

	defer func() {
		// make sure to recover from any app-related panics to allow proper socket cleanup
		r := recover()
		if r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err := fmt.Errorf("recovered from panic: %v\n%s", r, buf)
			if !s.isLoggerSet {
				fmt.Fprintln(os.Stderr, err)
			}
			closeConn <- err
			s.appMtx.Unlock()
		}
	}()

	for {

		var req = &cmtabci.Request{}
		err := cmtabci.ReadMessage(bufReader, req)
		if err != nil {
			if err == io.EOF {
				closeConn <- err
			} else {
				closeConn <- fmt.Errorf("error reading message: %w", err)
			}
			return
		}
		s.appMtx.Lock()
		count++
		s.handleRequest(req, responses)
		s.appMtx.Unlock()
	}
}

func (s *SocketServer) handleRequest(req *cmtabci.Request, responses chan<- *cmtabci.Response) {
	switch r := req.Value.(type) {
	case *cmtabci.Request_Flush:
		responses <- &cmtabci.Response{Value: &cmtabci.Response_Flush{Flush: &cmtabci.ResponseFlush{}}}
	case *cmtabci.Request_Info:
		res, _ := s.app.Info(context.Background(), r.Info)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_Info{Info: res}}
	case *cmtabci.Request_CheckTx:
		res, _ := s.app.CheckTx(context.Background(), r.CheckTx)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_CheckTx{CheckTx: res}}
	case *cmtabci.Request_Commit:
		res, _ := s.app.Commit(context.Background(), r.Commit)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_Commit{Commit: res}}
	case *cmtabci.Request_Query:
		res, _ := s.app.Query(context.Background(), r.Query)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_Query{Query: res}}
	case *cmtabci.Request_InitChain:
		res, _ := s.app.InitChain(context.Background(), r.InitChain)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_InitChain{InitChain: res}}
	case *cmtabci.Request_ListSnapshots:
		res, _ := s.app.ListSnapshots(context.Background(), r.ListSnapshots)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_ListSnapshots{ListSnapshots: res}}
	case *cmtabci.Request_OfferSnapshot:
		res, _ := s.app.OfferSnapshot(context.Background(), r.OfferSnapshot)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_OfferSnapshot{OfferSnapshot: res}}
	case *cmtabci.Request_LoadSnapshotChunk:
		res, _ := s.app.LoadSnapshotChunk(context.Background(), r.LoadSnapshotChunk)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_LoadSnapshotChunk{LoadSnapshotChunk: res}}
	case *cmtabci.Request_ApplySnapshotChunk:
		res, _ := s.app.ApplySnapshotChunk(context.Background(), r.ApplySnapshotChunk)
		responses <- &cmtabci.Response{Value: &cmtabci.Response_ApplySnapshotChunk{ApplySnapshotChunk: res}}
	default:
		responses <- &cmtabci.Response{Value: &cmtabci.Response_Exception{Exception: &cmtabci.ResponseException{Error: "Unknown request"}}}
	}
}

// Pull responses from 'responses' and write them to conn.
func (s *SocketServer) handleResponses(closeConn chan error, conn io.Writer, responses <-chan *cmtabci.Response) {
	var count int
	var bufWriter = bufio.NewWriter(conn)
	for {
		var res = <-responses
		err := cmtabci.WriteMessage(res, bufWriter)
		if err != nil {
			closeConn <- fmt.Errorf("error writing message: %w", err)
			return
		}
		if _, ok := res.Value.(*cmtabci.Response_Flush); ok {
			err = bufWriter.Flush()
			if err != nil {
				closeConn <- fmt.Errorf("error flushing write buffer: %w", err)
				return
			}
		}
		count++
	}
}
