package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type PostingWName struct {
	Filename string
	Count    int64
}

func MakeCode(code int64) []byte {
	c := make([]byte, 8)
	binary.PutVarint(c, code)
	return c
}

func main() {
	query := []string{"again", "myself", "judge"}

	c, err := net.Dial("tcp4", ":8000")
	if err != nil {
		log.Panicf("error connecting to server: %s", err)
	}
	fmt.Println("connected to server")
	for _, q := range query {
		var cod int64
		var l int64
		buf := bytes.Buffer{}
		buf.Write(MakeCode(int64(len(q))))
		buf.WriteString(q)
		_, err := c.Write(buf.Bytes())
		if err != nil {
			log.Panicf("error sending query to server: %s", err)
		}
		fmt.Println("sent request:", q)
		code := make([]byte, 8)
		_, err = c.Read(code)
		if err != nil {
			log.Panicf("error getting response from server: %s", err)
		}
		cod, _ = binary.Varint(code)
		fmt.Println("got response:", cod)
		if cod != 200 {
			log.Printf("error occured during query '%s', code: %d\n", q, cod)
			continue
		}
		_, err = c.Read(code)
		if err != nil {
			log.Panicf("error getting length of body: %s", err)
		}
		l, _ = binary.Varint(code)
		if l == 0 {
			log.Panicf("response is 0 bytes")
		}
		body := make([]byte, l)
		_, err = c.Read(body)
		if err != nil {
			log.Panicf("error getting body: %s", err)
		}
		res := []PostingWName{}
		err = json.Unmarshal(body, &res)
		if err != nil {
			log.Panicf("error unmarshaling body: %s", err)
		}
		fmt.Println(res, "\n")
	}

}
