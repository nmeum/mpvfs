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
	msgID int32

	respMtx *sync.Mutex
	respMap map[int32]chan response

	propMtx   *sync.Mutex
	propID    int32
	propChans map[int32]chan interface{}

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
		msgID:     math.MinInt32,
		propID:    math.MinInt32,
		respMtx:   new(sync.Mutex),
		respMap:   make(map[int32]chan response),
		propMtx:   new(sync.Mutex),
		propChans: make(map[int32]chan interface{}),
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

func (c *Client) nextID() int32 {
	// XXX: Mutex needed here?
	if c.msgID == math.MaxInt32 {
		c.msgID = math.MinInt32
	} else {
		c.msgID++
	}

	return c.msgID
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
	id := c.propID
	c.propChans[id] = ch

	// TODO: Don't reuse existing property IDs on overflow.
	c.propID++
	c.propMtx.Unlock()

	_, err := c.ExecCmd("observe_property", id, name)
	if err != nil {
		c.propMtx.Lock()
		delete(c.propChans, id)
		c.propMtx.Unlock()
		return nil, err
	}

	return ch, nil
}
