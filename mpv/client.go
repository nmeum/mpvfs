package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"errors"
)

type Client struct {
	id msgID
	mq queue

	conn    net.Conn
	msgChan chan response
}

func NewClient(path string) (*Client, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}

	c := &Client{
		id:      math.MinInt32,
		mq:      newQueue(),
		conn:    conn,
		msgChan: make(chan response),
	}

	go c.recvLoop(c.msgChan)
	go c.dispatchLoop(c.msgChan)

	return c, nil
}

func (c *Client) recvLoop(ch chan<- response) {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		fmt.Println("data:", string(data))

		var msg response
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

func (c *Client) dispatchLoop(ch <-chan response) {
	for {
		msg := <-ch
		c.mq.Signal(msg)
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

func (c *Client) ExecCmd(name string, args ...string) (interface{}, error) {
	cmd := c.newCmd(name, args...)
	err := cmd.Encode(c.conn)
	if err != nil {
		return nil, err
	}

	// Every message must be terminated with \n.
	_, err = c.conn.Write([]byte("\n"))
	if err != nil {
		return nil, err
	}

	response := c.mq.Wait(cmd.ID)
	if response.Error != "success" {
		return nil, errors.New(response.Error)
	}

	return response.Data, nil
}
