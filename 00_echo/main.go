package main

import (
	"io"
	"log"
	"net"
)

const maxConnections = 10

func main() {
	l, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	slots := make(chan struct{}, maxConnections)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		slots <- struct{}{}

		go func(c net.Conn) {
			defer func() {
				log.Printf("Closing connection from %s", c.RemoteAddr())
				<-slots
				c.Close()
			}()

			log.Printf("Accepted connection from %s", c.RemoteAddr())
			written, rerr := io.Copy(c, c)
			if rerr != nil {
				log.Printf("Echo error from %s: %v", c.RemoteAddr(), rerr)
				return
			}

			log.Printf("Echoed %d bytes from %s", written, c.RemoteAddr())
		}(conn)
	}
}
