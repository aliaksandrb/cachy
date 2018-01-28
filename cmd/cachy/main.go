package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aliaksandrb/cachy/server"
)

func main() {
	s, err := server.New(server.MemoryStore)
	if err != nil {
		panic(err)
	}

	s.Set("yy", 1, 5*time.Second)
	s.Set("gg", "xxx", 5*time.Second)

	for i := 0; i < 100; i++ {
		go func(x int) {
			time.Sleep(time.Duration(rand.Intn(x+10)) * time.Millisecond)
			s.Set(fmt.Sprintf("%d", x), x, time.Duration(rand.Intn(x+1))*time.Millisecond)
		}(i)

		go func(x int) {
			time.Sleep(time.Duration(rand.Intn(x+100)) * time.Millisecond)
			fmt.Println(s.Get(fmt.Sprintf("%d", x)))
		}(i)
	}

	for {
		fmt.Println(s.Get("gg"))
		fmt.Println("------------------")
		//fmt.Println("-----", s, "-----------")
		time.Sleep(3 * time.Second)

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
