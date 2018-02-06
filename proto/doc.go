/*
Package proto provides primitives and interfaces for encoding/decoding
incomming/outcomming messages and runtime objects according to protocol format.

Protocol definition:

| Command | Leading Byte |
|---------|--------------|
| GET     | #            |
| SET     | +            |
| UPDATE  | ^            |
| REMOVE  | -            |
| KEYS    | ~            |


|        Runtime types        | Leading Byte |
|-----------------------------|--------------|
| error                       | !            |
| string                      | $            |
| int                         | &            |
| []interface{}               | @            |
| map[interface{}]interface{} | :            |
| nil                         | ~            |


|  Escape chars  | Byte |
|----------------|------|
| end of message | \r   |
| segment escape | \n   |

Some simple rules to follow:
- message parts (segments) are separated by "segment escape"
- all incomming value strings are equoted
- keys, ttl, map/slice sizes are not encoded, only values
- the size of map/slice follows its leading byte
- elements of map/slice follows its head, one by one, separated by "segment escape"
- messages contain only leading byte considered as nil of the type

Examples:

|                             Action                             |                      Message                  |
|----------------------------------------------------------------|-----------------------------------------------|
| SET some_key value "value" with ttl 1000                       | +\nsome_key\n$\"value\"\n1000\r							 |
| GET some_key                                                   | #\nsome_key\r                                 |
| UPDATE some_key to []interface{100, "cool\tstory"} without ttl | ^\nsome_key\n@2\n&100\n$\"cool\\tstory\"\n0\r |
| REMOVE some_key                                                | -\nsome_key\r                                 |
| KEYS                                                           | ~\n\r                                         |

More examples could be found in decoder_test.go and encoder_test.go.
*/

package proto
