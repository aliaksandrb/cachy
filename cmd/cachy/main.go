package main

import (
	"github.com/aliaksandrb/cachy/server"
	"github.com/aliaksandrb/cachy/store"
)

func main() {
	// TODO pass as config?
	// TODO probbly panic in Run
	err := server.Run(store.MemoryStore, "tcp", ":3000")
	if err != nil {
		panic(err)
	}

	//	go s.Set("yy", 1, 5*time.Second)
	//	go s.Set("gg", "xxx", 5*time.Second)
	//	fmt.Println(s)
	//
	//	go s.Get("yy")
	//	go s.Get("gg")
	//	time.Sleep(3 * time.Second)
	//	go s.Update("gg", "ggggggg", 1*time.Second)
	//	fmt.Println(s)
	//	go s.Get("yy")
	//	go s.Get("gg")
	//
	//	fmt.Println(s)

	///	for i := 0; i < 1000; i++ {
	///		go func(x int) {
	///			time.Sleep(time.Duration(rand.Intn(x+10)) * time.Millisecond)
	///			s.Set(fmt.Sprintf("%d", x), x, time.Duration(rand.Intn(x+1))*time.Millisecond)
	///		}(i)
	///
	///		go func(x int) {
	///			for {
	///				time.Sleep(time.Duration(rand.Intn(x+100)) * time.Millisecond)
	///				s.Get(fmt.Sprintf("%d", x))
	///			}
	///		}(i)
	///
	///		go func(x int) {
	///			time.Sleep(time.Duration(10 * time.Second))
	///			s.Remove(fmt.Sprintf("%d", x))
	///		}(i)
	///
	///	}
	///
	///	for {
	///		fmt.Println(s.Get("gg"))
	///		fmt.Println("------------------")
	///		time.Sleep(3 * time.Second)
	///		fmt.Println(s)
	///
	///	}
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
