package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
)

const REQUEST_LENGTH = 9
const INSERT = byte('I')
const QUERY = byte('Q')

type Operation struct {
	Type    byte
	Number1 int32
	Number2 int32
}

type Request struct {
	Dates []int
	Value []int
	ID    string
	buf   []byte
}

type OpResult struct {
	IsQuery bool
	Done    bool
	Value   int32
}

func main() {

	l, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			defer func() {
				log.Printf("Closing connection from %s", c.RemoteAddr())
				c.Close()
			}()
			req := NewRequest(c.RemoteAddr().String())

			for {
				res := req.Read(c)

				if res.Done {
					return
				}

				if res.IsQuery {
					rep := toResponse(res.Value)
					c.Write(rep)
				}
			}
		}(conn)
	}
}

func NewRequest(addr string) *Request {
	return &Request{
		Dates: make([]int, 0),
		Value: make([]int, 0),
		ID:    addr,
		buf:   make([]byte, REQUEST_LENGTH),
	}
}

func (req *Request) Read(c io.Reader) OpResult {
	_, err := io.ReadAtLeast(c, req.buf, REQUEST_LENGTH)

	if err != nil {
		log.Printf("read error from %s: %v", req.ID, err)
		return OpResult{Done: true}
	}
	op, err2 := NewOperation(req.ID, bytes.NewBuffer(req.buf))
	if err2 != nil {
		log.Printf("operation error from %s: %v", req.ID, err2)
		return OpResult{Done: true}
	}

	res, err3 := req.Process(op)
	if err3 != nil {
		log.Printf("process error from %s: %v", req.ID, err3)
		return OpResult{Done: true}
	}

	return res
}

func NewOperation(addr string, buf *bytes.Buffer) (*Operation, error) {
	var op Operation
	err := binary.Read(buf, binary.BigEndian, &op)
	if err != nil {
		return nil, fmt.Errorf("failed to read operation: %v", err)
	}
	return &op, nil
}

func (req *Request) Process(op *Operation) (OpResult, error) {
	switch op.Type {
	case INSERT:
		log.Printf("Insert operation: %d", op.Number1)
		return req.InsertData(int(op.Number1), int(op.Number2))
	case QUERY:
		log.Printf("Query operation: %d to %d", op.Number1, op.Number2)
		return req.QueryData(int(op.Number1), int(op.Number2))
	default:
		log.Printf("Unknown operation type: %d", op.Type)
	}
	return OpResult{Done: true}, nil
}

func (req *Request) InsertData(date int, value int) (OpResult, error) {
	idx := sort.SearchInts(req.Dates, date)
	req.Dates = append(req.Dates, 0)
	req.Value = append(req.Value, 0)

	copy(req.Dates[idx+1:], req.Dates[idx:])
	req.Dates[idx] = date
	copy(req.Value[idx+1:], req.Value[idx:])
	req.Value[idx] = value
	log.Printf("Inserted data: date=%d, value=%d", date, value)

	return OpResult{IsQuery: false}, nil
}

func (req *Request) QueryData(date1 int, date2 int) (OpResult, error) {
	if len(req.Dates) == 0 {
		log.Printf("No data available for query: date1=%d, date2=%d", date1, date2)
		return OpResult{IsQuery: true, Value: 0}, nil
	}

	startIdx := sort.SearchInts(req.Dates, date1)
	endIdx := sort.SearchInts(req.Dates, date2+1)

	if startIdx >= len(req.Dates) || endIdx <= startIdx {
		log.Printf("No data in range for query: date1=%d, date2=%d", date1, date2)
		return OpResult{IsQuery: true, Value: 0}, nil
	}

	sum := 0
	count := 0
	for i := startIdx; i < endIdx; i++ {
		sum += req.Value[i]
		count++
	}

	if count == 0 {
		log.Printf("No data in range for query: date1=%d, date2=%d", date1, date2)
		return OpResult{IsQuery: true, Value: 0}, nil
	}

	mean := sum / count
	log.Printf("Query result: mean=%d for range date1=%d to date2=%d", mean, date1, date2)
	return OpResult{IsQuery: true, Value: int32(mean)}, nil
}

func processRequest(req []byte) (int32, error) {
	if len(req) != REQUEST_LENGTH {
		return 0, fmt.Errorf("invalid request length: %d", len(req))
	}
	return 0, nil
}

func toResponse(n int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, n)
	return buf.Bytes()
}
