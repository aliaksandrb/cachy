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
	fmt.Printf("new server is running on %s localhost%s\n", p, service)

	for {
		client, err := listener.Accept()
		if err != nil {
			fmt.Println("client connection error: ", err)
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

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadBytes('\r')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Println("read error: ", err)
			return
		}

		resp, err := s.processMsg(msg)
		if err != nil {
			fmt.Println("processing msg error: ", err)
			return
		}

		_, err = conn.Write(resp)
		if err != nil {
			fmt.Println("write error: ", err)
			return
			// TODO retry
		}
	}
}

type server struct {
	store store.Store
}

func (s *server) processMsg(m []byte) ([]byte, error) {
	log.Info("\nstore %+v\n", s.store)

	fmt.Printf("got: %q\n", m)

	//TODO bad place ehere
	decoder := proto.NewDecoder()

	parsed, err := decoder.Decode(bytes.NewReader(m))
	if err != nil {
		return proto.Encode(err, false)
	}

	var v interface{}

	// TODO what is herhe?
	if r, ok := parsed.(*proto.Request); ok {
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
