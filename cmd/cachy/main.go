package main

import (
	"flag"

	"github.com/aliaksandrb/cachy/server"
)

func main() {
	bSize := flag.Int("bsize", 32, "how many buckets to use, default: 32")
	port := flag.String("port", "3000", "port number to run, default: 3000")
	flag.Parse()

	s, err := server.Run(server.MemoryStore, *bSize, ":"+*port)
	if err != nil {
		panic(err)
	}
	<-s.Done()
}
