package mpv

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"sync"
)

type Client struct {
	id msgID
	mq queue

	propMtx   *sync.Mutex
	propChans map[string]chan interface{}

	conn    net.Conn
	msgChan chan response

	ErrChan <-chan error
}

func NewClient(path string) (*Client, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}

	c := &Client{
		id:        math.MinInt32,
		mq:        newQueue(),
		propMtx:   new(sync.Mutex),
		propChans: make(map[string]chan interface{}),
		conn:      conn,
		msgChan:   make(chan response),
	}

	errChan := make(chan error)
	c.ErrChan = errChan

	go c.recvLoop(c.msgChan, errChan)
	go c.dispatchLoop(c.msgChan)

	return c, nil
}

func (c *Client) recvLoop(ch chan<- response, errCh chan<- error) {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		fmt.Println("data:", string(data))

		var msg response
		err := json.Unmarshal(data, &msg)
		if err != nil {
			errCh <- err
			continue
		}

		ch <- msg
	}

	err := scanner.Err()
	if err != nil {
		errCh <- err
	}
}

func (c *Client) dispatchLoop(ch <-chan response) {
	for {
		msg := <-ch
		if msg.Event == "property-change" {
			c.handleChange(msg)
		} else {
			c.mq.Signal(msg)
		}
	}
}

func (c *Client) handleChange(msg response) {
	c.propMtx.Lock()
	ch, ok := c.propChans[msg.PropertyName]
	if ok {
		ch <- msg.Data
	}
	c.propMtx.Unlock()
}

func (c *Client) nextID() msgID {
	if c.id == math.MaxInt32 {
		c.id = math.MinInt32
	} else {
		c.id++
	}

	return c.id
}

func (c *Client) newReq(name interface{}, args ...interface{}) *request {
	argv := append([]interface{}{name}, args...)
	return &request{Cmd: argv, ID: c.nextID()}
}

func (c *Client) ExecCmd(name string, args ...interface{}) (interface{}, error) {
	req := c.newReq(name, args...)
	err := req.Encode(c.conn)
	if err != nil {
		return nil, err
	}

	// Every message must be terminated with \n.
	_, err = c.conn.Write([]byte("\n"))
	if err != nil {
		return nil, err
	}

	response := c.mq.Wait(req.ID)
	if response.Error != noError {
		return nil, errors.New(response.Error)
	}

	return response.Data, nil
}

func (c *Client) SetProperty(name string, value interface{}) error {
	_, err := c.ExecCmd("set_property", name, value)
	return err
}

func (c *Client) GetProperty(name string) (interface{}, error) {
	value, err := c.ExecCmd("get_property", name)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// TODO: Allow multiple observers of same property
func (c *Client) ObserveProperty(name string) (<-chan interface{}, error) {
	c.propMtx.Lock()
	ch, ok := c.propChans[name]
	if ok {
		return nil, errors.New("property already observed")
	}

	ch = make(chan interface{})
	c.propChans[name] = ch

	_, err := c.ExecCmd("observe_property", 1, name)
	if err != nil {
		delete(c.propChans, name)
	}
	c.propMtx.Unlock()

	return ch, nil
}
