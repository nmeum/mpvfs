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

	respMtx *sync.Mutex
	respMap map[msgID]chan response

	propMtx   *sync.Mutex
	propId    msgID
	propChans map[msgID]chan interface{}

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
		propId:    math.MinInt32,
		respMtx:   new(sync.Mutex),
		respMap:   make(map[msgID]chan response),
		propMtx:   new(sync.Mutex),
		propChans: make(map[msgID]chan interface{}),
		conn:      conn,
		msgChan:   make(chan response, 5),
	}

	errChan := make(chan error, 1)
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
			go c.handleChange(msg)
		} else {
			go c.handleResp(msg)
		}
	}
}

func (c *Client) handleResp(msg response) {
	c.respMtx.Lock()
	ch, ok := c.respMap[msg.ID]
	c.respMtx.Unlock()

	if ok {
		ch <- msg
	}
}

func (c *Client) handleChange(msg response) {
	c.propMtx.Lock()
	ch, ok := c.propChans[msg.ID]
	c.propMtx.Unlock()

	if ok {
		ch <- msg.Data
	}
}

func (c *Client) nextID() msgID {
	// XXX: Mutex needed here?
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

	ch := make(chan response)
	defer close(ch)

	c.respMtx.Lock()
	c.respMap[req.ID] = ch
	defer delete(c.respMap, req.ID)
	c.respMtx.Unlock()

	// Every message must be terminated with \n.
	_, err = c.conn.Write([]byte("\n"))
	if err != nil {
		return nil, err
	}

	resp := <-ch
	if resp.Error != noError {
		return nil, errors.New(resp.Error)
	}

	return resp.Data, nil
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

func (c *Client) ObserveProperty(name string) (<-chan interface{}, error) {
	c.propMtx.Lock()
	ch := make(chan interface{})
	id := c.propId
	c.propChans[id] = ch

	// TODO: Don't reuse existing property IDs on overflow.
	c.propId += 1
	c.propMtx.Unlock()

	_, err := c.ExecCmd("observe_property", id, name)
	if err != nil {
		delete(c.propChans, id)
	}

	return ch, nil
}
