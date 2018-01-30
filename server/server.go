package server

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/aliaksandrb/cachy/proto"
	"github.com/aliaksandrb/cachy/store"
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
		msg, err := reader.ReadBytes('\n')
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
			//return
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
	fmt.Printf("got: %q\n", m)

	v, err := proto.Decode(m)
	if err != nil {
		return proto.Encode(err, false)
	}

	out, err := proto.Encode(v, false)
	fmt.Printf("out: %q\n", out)
	if err != nil {
		return proto.Encode(err, false)
	}

	return out, nil
}
