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

func Itbs(code int64) []byte {
	c := make([]byte, 8)
	binary.PutVarint(c, code)
	return c
}

func ReadBytes(r io.Reader, n int64) ([]byte, error) {
	ba := make([]byte, n)
	count, err := r.Read(ba)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes: %w", err)
	}
	if int64(count) != n {
		return nil, fmt.Errorf("not enough bytes")
	}
	return ba, nil
}

func ReadInt64(r io.Reader) (int64, error) {
	ba := make([]byte, 8)
	n, err := r.Read(ba)
	if err != nil {
		return 0, fmt.Errorf("error reading bytes: %w", err)
	}
	if n != 8 {
		return 0, fmt.Errorf("not enough bytes")
	}
	res, _ := binary.Varint(ba)
	return res, nil
}
func WriteRequest(wr io.Writer, rq string) error {
	rqb := make([]byte, 8, 8+len(rq))
	binary.PutVarint(rqb[0:8], int64(len(rq)))
	rqb = append(rqb, []byte(rq)...)
	_, err := wr.Write(rqb)
	if err != nil {
		return fmt.Errorf("error writing request: %w", err)
	}
	return nil
}

type Response struct {
	Code   int64
	Length int64
	Body   []byte
}

func ReadResponse(r io.Reader) (*Response, error) {
	var resp Response
	var err error
	resp.Code, err = ReadInt64(r)
	if err != nil {
		return nil, fmt.Errorf("error reading code: %w", err)
	}
	resp.Length, err = ReadInt64(r)
	if err != nil {
		return nil, fmt.Errorf("error reading body length: %w", err)
	}
	resp.Body, err = ReadBytes(r, resp.Length)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	return &resp, nil
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
