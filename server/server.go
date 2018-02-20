package server

import (
	"bufio"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aliaksandrb/cachy/proto"
	"github.com/aliaksandrb/cachy/store"
	"github.com/aliaksandrb/cachy/store/mstore"

	log "github.com/aliaksandrb/cachy/logger"
)

// Server represents a basic storage servier.
type Server interface {
	Stop() error
	Done() <-chan struct{}
}

// Run spawns and runs a new server in background.
func Run(s storeType, bs int, addr string) (Server, error) {
	listener, err := makeListener(addr)
	if err != nil {
		return nil, err
	}

	server, err := New(s, bs, listener)
	if err != nil {
		return nil, err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signals
		log.Info("shutdown")
		server.Stop()
	}()

	log.Info("server started on %s ...", addr)
	go server.start()

	return server, nil
}

// Stop attempts to gracefully stop the server by stopping accept any new connections and let existing ones to finish.
// After a timeout it exits anyway.
func (s *server) Stop() error {
	log.Info("stoping server gracefully...")
	defer func() {
		close(s.done)
	}()

	close(s.closing)

	select {
	case <-s.syncClients():
	case <-time.After(10 * time.Second):
		log.Info("timeouted, killing...")
	}

	if err := s.listener.Close(); err != nil {
		return err
	}

	log.Info("stoping server, done.")
	return nil
}

// Done notifies if the server being stopped.
func (s *server) Done() <-chan struct{} {
	return s.done
}

func (s *server) syncClients() chan struct{} {
	done := make(chan struct{})
	go func() {
		s.clients.Wait()
		close(done)
	}()

	return done
}

type storeType int

const (
	MemoryStore = storeType(iota)
	PersistantStore
)

// New returns a new Server implementation.
func New(s storeType, bs int, l *net.TCPListener) (*server, error) {
	var db store.Store

	if s != MemoryStore {
		return nil, store.ErrUnsuportedStoreType
	}

	db, err := mstore.New(bs, 0)
	if err != nil {
		return nil, err
	}

	return &server{
		store:    db,
		listener: l,
		closing:  make(chan struct{}),
		done:     make(chan struct{}),
		clients:  &sync.WaitGroup{},
		decoder:  proto.NewDecoder(),
		writer:   proto.NewWriter(),
	}, nil
}

type server struct {
	store    store.Store
	closing  chan struct{}
	done     chan struct{}
	clients  *sync.WaitGroup
	listener *net.TCPListener
	decoder  MessageDecoder
	writer   Writer
}

// MessageDecoder used to decode incomming TCP messages on a server side.
type MessageDecoder interface {
	// DecodeMessage decodes message in a buf into message m.
	// Currently it is used only on server side.
	// It should never panic because of user input.
	DecodeMessage(buf *bufio.Reader) (m interface{}, err error)
}

// Writer used to write streams of encoded data into io.Writer.
type Writer interface {
	// Write encodes obj and writes data into io.Writer w.
	// It fails only if it ca not write to w.
	Write(w io.Writer, obj interface{}) error
	// WriteRaw writes bytes as it is.
	WriteRaw(w io.Writer, b []byte) error
	// WriteUnknownErr is generally a shortcut to Write(w, ErrUknown).
	WriteUnknownErr(w io.Writer) error
}

func (s *server) start() {
	for {
		select {
		case <-s.closing:
			return
		default:
			client, err := s.listener.Accept()
			if err != nil {
				log.Err("client connection error: %v", err)
				continue
			}

			s.clients.Add(1)
			go s.handleClient(client)
		}
	}
}

func (s *server) handleClient(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Err("handle client error: %v", err)
		}
	}()
	defer s.clients.Done()
	defer conn.Close()

	log.Info("new client connected: %+v", conn.RemoteAddr())

	var err error
	reader := bufio.NewReader(conn)

	for {
		select {
		case <-s.closing:
			return
		default:
			if err = s.handleMessage(reader, conn); err != nil {
				log.Info("closing a client: %+v", conn.RemoteAddr())
				return
			}
			reader.Reset(conn)
		}
	}
}
func (s *server) handleMessage(buf *bufio.Reader, w io.Writer) error {
	msg, err := s.decoder.DecodeMessage(buf)
	if err == io.EOF {
		return err
	}

	if err != nil {
		return s.writer.Write(w, err)
	}

	req, ok := msg.(*proto.Req)
	if !ok {
		log.Err("unknown message type: %q", msg)
		return s.writer.WriteUnknownErr(w)
	}

	result, err := s.processRequest(req)
	if err != nil {
		return s.writer.Write(w, err)
	}

	return s.writer.WriteRaw(w, result)
}

func (s *server) processRequest(r *proto.Req) (v []byte, err error) {
	switch r.Cmd {
	case proto.CmdGet:
		return s.store.Get(r.Key)
	case proto.CmdSet:
		return nil, s.store.Set(r.Key, r.Value, r.TTL)
	case proto.CmdUpdate:
		return nil, s.store.Update(r.Key, r.Value, r.TTL)
	case proto.CmdRemove:
		return nil, s.store.Remove(r.Key)
	case proto.CmdKeys:
		return proto.Encode(s.store.Keys())
	}

	return nil, proto.ErrUnknown
}

func makeListener(addr string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return net.ListenTCP("tcp", tcpAddr)
}
