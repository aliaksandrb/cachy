# POC

In-memory store, bucketing using Golang's hashes (not clever enough).

## Install

```bash
go get github.com/aliaksandrb/cachy/...
```

## Run

You can run the executable with two flags:

- `-bsize` : sets the buckets number for underlaying storage (default: 32)
- `-port` : sets the port number to listen for requests (default: 3000)

Example:

```bash
$GOPATH/bin/cachy -bsize 100 -port 4000
```

To stop just terminate it using `^C`.

## Usage

Assuming server is running on the same machine using port 3000,
here is an example script that will store, get and delete some values.

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aliaksandrb/cachy/client"
)

func main() {
	session, err := client.New("127.0.0.1:3000", 5) // Where 5 is connections pull for concurency. By default it is 1 which means requests are queued.
	checkError(err)
	defer session.Close() // Don't forget to Close it when done.

	err = session.Set("test", "value", 0)
	checkError(err)

	val, err := session.Get("test")
	checkError(err)
	fmt.Println(val) // "value"

	keys, err := session.Keys()
	checkError(err)
	fmt.Println(keys) // ["test"]

	err = session.Remove("test")
	checkError(err)

	_, err = session.Get("test")
	fmt.Println(err) // "not found"
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	}
}

```

## API Reference

Here is the list of methods available for the client:

- Get(key string) (val interface{}, err error)
- Set(key string, val interface{}, ttl time.Duration) error
- Update(key string, val interface{}, ttl time.Duration) error
- Remove(key string) error
- Keys() ([]string, error)
- Close()

Where concrete value of `val` supposed to be one from the following list:
- string
- int
- nil
- []interface{}
- map[interface{}]interface{}

## License

MIT

You won't use it anyway :)
