package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"math"
	"math/big"
	"net"
)

type TCPRequest struct {
	Method string  `json:"method" required:"true"`
	Number float64 `json:"number" required:"true"`
}

type TCPResponse struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func hasFractionalPart(f float64) bool {
	return f != math.Trunc(f)
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

			reader := bufio.NewReader(c)

			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					// not eof means that some other error happened and we should return malformed response and close connection
					if err != io.EOF {
						log.Printf("read error from %s: %v", c.RemoteAddr(), err)
						c.Write([]byte(`{"error": "read error"}` + "\n"))
						return
					}
				}

				req := TCPRequest{}
				var raw map[string]json.RawMessage
				jerr := json.Unmarshal(line, &raw)

				if jerr == nil {
					_, hasMethod := raw["method"]
					_, hasNumber := raw["number"]

					if !hasMethod || !hasNumber {
						log.Printf("missing required fields from %s", c.RemoteAddr())
						c.Write([]byte(`{"error": "missing required fields"}` + "\n"))
						return
					}

					jerr = json.Unmarshal(line, &req)
				}

				// Invalid JSON malformed request and response
				if jerr != nil {
					log.Printf("json unmarshal error from %s: %v", c.RemoteAddr(), jerr)
					c.Write([]byte(`{"error": "invalid json"}` + "\n"))
					return
				}

				// Invalid value for method field -> malformed response
				if req.Method != "isPrime" {
					log.Printf("invalid method from %s: %s", c.RemoteAddr(), req.Method)
					c.Write([]byte(`{"error": "invalid method"}` + "\n"))
					return
				}

				isPrime := !hasFractionalPart(req.Number) && big.NewInt(int64(req.Number)).ProbablyPrime(0) // 0 means default certainty

				resp := TCPResponse{
					Method: req.Method,
					Prime:  isPrime,
				}

				log.Printf("received from %s: %s", c.RemoteAddr(), line)
				respBytes, _ := json.Marshal(resp)
				c.Write(append(respBytes, '\n'))

				if err == io.EOF {
					log.Printf("connection closed by client %s", c.RemoteAddr())
					return
				}
			}

		}(conn)
	}
}
