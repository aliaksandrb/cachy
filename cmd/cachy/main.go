package main

import (
	"fmt"

	"github.com/aliaksandrb/cachy/server"
)

func main() {
	_, err := server.Run(server.MemoryStore, 32, ":3000")
	if err != nil {
		panic(err)
	}
	//time.Sleep(10 * time.Second)
	//defer s.Stop()
	fmt.Scanln()
}
