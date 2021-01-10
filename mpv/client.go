package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
)

type Client struct {
	id msgID

	conn    net.Conn
	msgChan chan message
}

func NewClient(path string) (*Client, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:    conn,
		id:      math.MinInt32,
		msgChan: make(chan message),
	}

	go c.recvLoop(c.msgChan)
	go c.dispatchLoop(c.msgChan)

	return c, nil
}

func (c *Client) recvLoop(ch chan<- message) {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		fmt.Println("data:", string(data))

		var msg message
		err := json.Unmarshal(data, &msg)
		if err != nil {
			panic(err) // XXX
		}

		ch <- msg
	}

	err := scanner.Err()
	if err != nil {
		panic(err) // XXX
	}
}

func (c *Client) dispatchLoop(ch <-chan message) {
	for {
		msg := <-ch
		if msg.isResponse() {
			fmt.Println("got response")
		} else {
			fmt.Println("unknown response")
		}
	}
}

func (c *Client) nextID() msgID {
	if c.id == math.MaxInt32 {
		c.id = math.MinInt32
	} else {
		c.id++
	}

	return c.id
}

func (c *Client) newCmd(name string, args ...string) *command {
	argv := append([]string{name}, args...)
	return &command{Cmd: argv, ID: c.nextID()}
}

func (c *Client) ExecCmd(name string, args ...string) error {
	cmd := c.newCmd(name, args...)
	err := cmd.Encode(c.conn)
	if err != nil {
		return err
	}

	// Every message must be terminated with \n.
	_, err = c.conn.Write([]byte("\n"))
	if err != nil {
		return err
	}

	// TODO
	return nil
}
