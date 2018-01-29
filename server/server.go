package server

import (
	"bufio"
	"bytes"
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

type server struct {
	store store.Store
}

func (s *server) processMsg(m []byte) ([]byte, error) {
	if len(m) == 0 {
		return nil, fmt.Errorf("empty message")
	}
	if m[0] == '\n' {
		m = m[1:]
	}

	fmt.Printf("\n===> Got: %q| \n %#v|\n", string(m), string(m))
	//TODO one interation
	m = bytes.Replace(bytes.Replace(m, []byte("\\n"), []byte{'\n'}, -1), []byte("\\r"), []byte{'\r'}, -1)

	v, err := proto.Decode(m[:len(m)-1])
	if err != nil {
		//TODO encode error
		return nil, err
	}
	fmt.Println("got: ", v)
	return []byte("hi from server\n"), nil
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
