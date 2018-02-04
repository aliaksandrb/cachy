package server

import (
	"bufio"
	"io"
	"net"

	"github.com/aliaksandrb/cachy/proto"
	"github.com/aliaksandrb/cachy/store"

	log "github.com/aliaksandrb/cachy/logger"
)

func Run(st store.Type, p string, service string) (err error) {
	tcpAddr, err := net.ResolveTCPAddr(p, service)
	if err != nil {
		return
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return
	}
	defer listener.Close()

	dbStore, err := store.New(st)
	if err != nil {
		return err
	}

	// TODO double server run?
	s := &server{store: dbStore}
	log.Info("new server is running on %s localhost%s\n", p, service)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Err("client connection error: %v", err)
			continue
		}

		go s.handleClient(client)
	}

	return nil
}

func (s *server) handleClient(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Err("handle client error: %v", err)
		}
	}()
	defer conn.Close()
	log.Info("new client connected: %+v", conn.RemoteAddr())

	var err error
	reader := bufio.NewReader(conn)

	for {
		if err = s.handleMessage(reader, conn); err != nil {
			log.Info("closing a client: %+v", conn.RemoteAddr())
			return
		}
		reader.Reset(conn)
	}
}

func (s *server) handleMessage(buf *bufio.Reader, w io.Writer) error {
	msg, err := proto.DecodeMessage(buf)
	if err == io.EOF {
		return err
	}
	if err != nil {
		return proto.Write(w, err)
	}

	req, ok := msg.(*proto.Req)
	if !ok {
		log.Err("unknown message type: %q", msg)
		return proto.WriteUnknownErr(w)
	}

	result, err := s.processRequest(req)
	if err != nil {
		return proto.Write(w, err)
	}

	return proto.Write(w, result)
}

type server struct {
	store store.Store
}

func (s *server) processRequest(r *proto.Req) (v interface{}, err error) {
	switch r.Cmd {
	case proto.CmdGet:
		v, _, err := s.store.Get(r.Key)
		return v, err
	case proto.CmdSet:
		return nil, s.store.Set(r.Key, r.Value, r.TTL)
	case proto.CmdUpdate:
		return nil, s.store.Update(r.Key, r.Value, r.TTL)
	case proto.CmdRemove:
		return nil, s.store.Remove(r.Key)
	case proto.CmdKeys:
		return s.store.Keys(), nil
	}

	return nil, proto.ErrUnknown
}
