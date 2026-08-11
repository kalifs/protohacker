package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

const (
	PORT            = ":8080"
	MAX_CONNECTIONS = 20
)

type Room struct {
	clients map[net.Conn]*Client
}

type Status byte

const (
	STATUS_CONNECTING Status = iota
	STATUS_CONNECTED
	STATUS_DISCONNECTED
)

var maxId int = 0

type Client struct {
	conn   net.Conn
	name   string
	status Status
	id     int
	reader *bufio.Scanner
}

func main() {
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	slots := make(chan struct{}, MAX_CONNECTIONS)
	room := NewRoom()

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
			room.HandleConnection(c)
		}(conn)
	}
}

func NewRoom() *Room {
	return &Room{
		clients: make(map[net.Conn]*Client),
	}
}

func (r *Room) HandleConnection(c net.Conn) {
	client, success := r.AddClient(c)
	if !success {
		return
	}
	r.SendMemberList(client)
	r.SendToOthers(client, "* "+client.name+" has entered the room")

	for {
		msg := client.GetMessage()
		if msg == "" {
			r.DisconnectClient(client)
			return
		}
		r.SendToOthers(client, "["+client.name+"]"+" "+msg)
	}
}

func (r *Room) AddClient(c net.Conn) (*Client, bool) {
	maxId++
	client := &Client{
		conn:   c,
		name:   "",
		status: STATUS_CONNECTING,
		id:     maxId,
		reader: bufio.NewScanner(c),
	}

	r.clients[c] = client
	log.Printf("Client joining %d from %s", client.id, c.RemoteAddr())
	success := client.GetName()
	if !success {
		delete(r.clients, c)
		return nil, false
	}
	client.status = STATUS_CONNECTED
	return client, true
}

func (r *Room) SendMemberList(c *Client) {
	var members []string
	for _, client := range r.clients {
		if client.status == STATUS_CONNECTED && client != c {
			members = append(members, client.name)
		}
	}

	c.SendMessage("* room members: " + strings.Join(members, ", "))
}

func (r *Room) SendToOthers(sender *Client, msg string) {
	for _, client := range r.clients {
		if client != sender && client.status == STATUS_CONNECTED {
			client.SendMessage(msg)
		}
	}
}

func (r *Room) DisconnectClient(c *Client) {
	c.status = STATUS_DISCONNECTED
	delete(r.clients, c.conn)
	r.SendToOthers(c, "* "+c.name+" has left the room")
}

func (c *Client) GetName() bool {
	c.SendMessage("What is your name?")
	pName := c.GetMessage()

	if len(pName) == 0 {
		log.Printf("Client %d from %s did not provide a name", c.id, c.conn.RemoteAddr())
		c.SendMessage("You did not provide a name. Disconnecting.")
		return false
	}

	if strings.ContainsFunc(pName, func(l rune) bool {
		if !(l >= 'a' && l <= 'z' || l >= 'A' && l <= 'Z' || l >= '0' && l <= '9') {
			return true
		}
		return false
	}) {
		log.Printf("Client %d from %s provided an invalid name", c.id, c.conn.RemoteAddr())
		c.SendMessage("Invalid name. Disconnecting.")
		return false
	}
	c.name = pName
	return true
}

func (c *Client) SendMessage(msg string) {
	_, err := c.conn.Write([]byte(msg + "\n"))
	if err != nil {
		log.Printf("Error writing to %s: %v", c.conn.RemoteAddr(), err)
	}
}

func (c *Client) GetMessage() string {
	hasErr := c.reader.Scan()

	if !hasErr {
		log.Printf("Error reading from %s: %v", c.conn.RemoteAddr(), c.reader.Err())
		return ""
	}

	line := c.reader.Text()
	return line
}
