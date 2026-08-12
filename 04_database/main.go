package main

import (
	"bytes"
	"log"
	"net"
	"os"
	"sync"
)

var host string

func init() {
	if os.Getenv("HOST") != "" {
		host = os.Getenv("HOST")
	} else {
		host = ""
	}
}

type DB struct {
	mutex  sync.Mutex
	values map[string]string
	sep    []byte
}

func main() {
	laddr, err := net.ResolveUDPAddr("udp", host+":5000")
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	db := NewDB()

	for {
		buf := make([]byte, 1000)
		n, addr, err := l.ReadFrom(buf)
		// // Ensure that size of data is never more than 1000 bytes.
		// if n >= 1000 {
		// 	n = 999
		// }
		if err != nil {
			log.Fatal(err)
		}

		go func(data []byte, addr net.Addr) {
			resp := db.processRequest(data)
			log.Printf("Sending response: %s", resp)
			if resp != "" {
				l.WriteTo([]byte(resp), addr)
			}
		}(buf[:n], addr)
	}
}

func NewDB() *DB {
	values := make(map[string]string)
	values["version"] = "Arturs Key-Value store v1.0"

	return &DB{
		values: values,
		sep:    []byte("="),
	}
}

func (db *DB) processRequest(data []byte) string {
	log.Printf("Received request: %s", string(data))

	before, after, found := bytes.Cut(data, db.sep)

	// before = bytes.TrimRight(before, "\n\r")
	// after = bytes.TrimRight(after, "\n\r")

	if !found {
		log.Printf("Getting value: %s", string(before))
		return db.GetKey(&before)
	}

	log.Printf("Setting value: %s=%s", string(before), string(after))
	return db.SetKey(&before, &after)
}

func (db *DB) SetKey(key *[]byte, value *[]byte) string {
	if string(*key) == "version" {
		return ""
	}
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.values[string(*key)] = string(*value)
	return ""
}

func (db *DB) GetKey(key *[]byte) string {
	value, exists := db.values[string(*key)]
	if !exists {
		return string(*key) + "="
	}
	return string(*key) + "=" + value
}
