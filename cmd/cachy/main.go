package main

import (
	"fmt"
	"time"

	"github.com/aliaksandrb/cachy/server"
)

func main() {
	s, err := server.New(server.MemoryStore)
	if err != nil {
		panic(err)
	}

	s.Set("yy", 1, 0)
	s.Set("gg", "xxx", time.Second)

	for {
		fmt.Println(s.Get("gg"))
		fmt.Println(s.Get("yy"))
		fmt.Println("-----", s, "-----------")
		time.Sleep(1 * time.Second)
	}
	//	v, err := proto.Decode([]byte(":2\r\n$key1\r\n$value1\r\n$key2\r\n$value2\r\n"))
	//	if err != nil {
	//		fmt.Println("ERROR: ", err)
	//	} else {
	//
	//		fmt.Printf("%v +++++ %T\n", v, v)
	//		x := v.(map[string]interface{})
	//		fmt.Printf("%v +++++ %T\n", x["key1"], x)
	//	}
}
