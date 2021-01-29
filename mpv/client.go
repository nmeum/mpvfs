package mpv

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net"
	"sync"
	"sync/atomic"
)

type Client struct {
	msgID int32

	respMtx *sync.Mutex
	respMap map[int32]chan response

	propMtx   *sync.Mutex
	propID    int32
	propChans map[int32]chan interface{}

	conn    net.Conn
	reqChan chan request
	msgChan chan response

	verbose bool
	ErrChan <-chan error
}

func NewClient(path string, verbose bool) (*Client, error) {
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
		reqChan:   make(chan request),
		msgChan:   make(chan response),
		verbose:   verbose,
	}

	errChan := make(chan error, 5)
	c.ErrChan = errChan

	go c.sendLoop(c.reqChan, errChan)
	go c.recvLoop(c.msgChan, errChan)

	go c.dispatchLoop(c.msgChan)
	return c, nil
}

func (c *Client) sendLoop(ch <-chan request, errCh chan<- error) {
	for req := range ch {
		err := req.Encode(c.conn)
		if err != nil {
			errCh <- err
			continue
		}

		// Every message must be terminated with \n.
		_, err = c.conn.Write([]byte("\n"))
		if err != nil {
			errCh <- err
			continue
		}
	}
}

func (c *Client) recvLoop(ch chan<- response, errCh chan<- error) {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		c.debug("<-", string(data))

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
		} else if msg.Event == "" {
			go c.handleResp(msg)
		}
	}
}

func (c *Client) handleResp(msg response) {
	c.respMtx.Lock()
	ch, ok := c.respMap[msg.ReqID]
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

func (c *Client) debug(args ...interface{}) {
	const prefix = "[mpv client]"
	if c.verbose {
		argv := append([]interface{}{prefix}, args...)
		log.Println(argv...)
	}
}

func (c *Client) newReq(name interface{}, args ...interface{}) request {
	argv := append([]interface{}{name}, args...)

	// Signed integer overflow is well-defined in go.
	id := atomic.AddInt32(&c.msgID, 1)

	return request{Cmd: argv, ID: id}
}

func (c *Client) ExecCmd(name string, args ...interface{}) (interface{}, error) {
	req := c.newReq(name, args...)

	ch := make(chan response)
	defer close(ch)

	c.respMtx.Lock()
	c.respMap[req.ID] = ch
	c.respMtx.Unlock()

	// Delete request from respMap on return
	defer func() {
		c.respMtx.Lock()
		delete(c.respMap, req.ID)
		c.respMtx.Unlock()
	}()

	c.debug("->", req.String())
	c.reqChan <- req

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
	ch := make(chan interface{})

	c.propMtx.Lock()
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
