package main

import (
	"github.com/aliaksandrb/cachy/server"
	"github.com/aliaksandrb/cachy/store"
)

func main() {
	// TODO pass as config?
	// TODO probbly panic in Run
	// TODO move store out of here
	err := server.Run(store.MemoryStore, "tcp", ":3000")
	if err != nil {
		panic(err)
	}
}
