package main

import (
	"fmt"

	"github.com/aliaksandrb/cachy/server"
	"github.com/aliaksandrb/cachy/store"
)

func main() {
	// TODO pass as config?
	// TODO move store out of here
	_, err := server.Run(store.MemoryStore, ":3000")
	if err != nil {
		panic(err)
	}
	//time.Sleep(10 * time.Second)
	//defer s.Stop()
	fmt.Scanln()
}
