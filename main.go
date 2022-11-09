package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
	fmt.Println("code:", resp.Code, "len:", resp.Length)
	resp.Body, err = ReadBytes(r, resp.Length)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	return &resp, nil
}

var filename string

func main() {
	flag.StringVar(&filename, "i", "", "path to file with queries")
	flag.Parse()
	if filename == "" {
		panic("input path is not specified!")
	}
	query := make([]string, 0, 10)
	file, err := os.Open(filename)
	if err != nil {
		log.Panicf("error opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		query = append(query, scanner.Text())
	}

	c, err := net.Dial("tcp4", "host.docker.internal:8000")
	if err != nil {
		log.Panicf("error connecting to server: %s", err)
	}
	fmt.Println("connected to server")
	for _, q := range query {
		err := WriteRequest(c, q)
		if err != nil {
			log.Panicf("error writing request: %s", err)
		}
		fmt.Println("sent request:", q)
		resp, err := ReadResponse(c)
		if err != nil {
			log.Panicf("error reading response: %s", err)
		}
		fmt.Printf("got response: %d\n", resp.Code)
		if resp.Code != 200 {
			if resp.Code == 404 {
				fmt.Println("word is not present in index\n")
				continue
			}
			if resp.Code == 500 {
				log.Panic("internal server error")
			}
		}
		res := []PostingWName{}
		err = json.Unmarshal(resp.Body, &res)
		if err != nil {
			log.Printf("error unmarshaling resp body: %s\n", err)
			continue
		}
		fmt.Println(res, "\n")
	}

}
