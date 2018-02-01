package server

import (
	"bufio"
	"bytes"
	"fmt"
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
			fmt.Println("client error: ", err)
		}
	}()
	defer conn.Close()

	log.Info("new client connected: %+v", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		log.Info("wating for a data...")

		msg, err := reader.ReadBytes('\r')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Err("read error: %v", err)
			return
		}

		log.Info("processing data: %q", msg)

		resp, err := s.processMsg(msg)
		if err != nil {
			log.Err("processing msg error: %v", err)
			return
		}

		log.Info("attempt to write: %q", resp)

		_, err = conn.Write(resp)
		if err != nil {
			log.Err("write error: %v", err)
			return
			// TODO retry
		}

		reader.Reset(conn)
	}
}

type server struct {
	store store.Store
}

func (s *server) processMsg(m []byte) ([]byte, error) {
	log.Info("\nStore %+v\n", s.store)

	//TODO bad place ehere

	parsed, err := proto.Decode(bytes.NewReader(m))
	if err != nil {
		return proto.Encode(err, false)
	}

	var v interface{}

	// TODO what is herhe?
	if r, ok := parsed.(*proto.Request); ok {
		log.Info("request: %+v", r)
		v, err = s.processRequest(r)
		if err != nil {
			return proto.Encode(err, false)
		}
	} else {
		v = parsed
	}

	out, err := proto.Encode(v, false)
	if err != nil {
		x, _ := proto.Encode(err, false)
		fmt.Printf("out: %q\n", x)
		return proto.Encode(err, false)
	}

	fmt.Printf("out: %q\n", out)

	return out, nil
}

func (s *server) processRequest(r *proto.Request) (v interface{}, err error) {
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
