package main

import (
	"fmt"

	"github.com/aliaksandrb/cachy/proto"
)

func main() {
	v, err := proto.Decode([]byte(":2\r\n$key1\r\n$value1\r\n$key2\r\n$value2\r\n"))
	if err != nil {
		fmt.Println("ERROR: ", err)
	} else {

		fmt.Printf("%v +++++ %T\n", v, v)
		x := v.(map[string]interface{})
		fmt.Printf("%v +++++ %T\n", x, x)
	}
}
