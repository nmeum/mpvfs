package mpv

import (
	"sync"
	"fmt"
)

type queue struct {
	lock *sync.Mutex
	msgs map[msgID]chan response
}

func newQueue() queue {
	return queue{
		lock: new(sync.Mutex),
		msgs: make(map[msgID]chan response),
	}
}

func (q *queue) getChan(id msgID) chan response {
	var ch chan response
	var ok bool

	q.lock.Lock()
	ch, ok = q.msgs[id]
	if !ok {
		ch = make(chan response)
		q.msgs[id] = ch
	}
	q.lock.Unlock()

	return ch
}

func (q *queue) Wait(id msgID) response {
	ch := q.getChan(id)
	fmt.Println("blocking reading", id)
	response := <-ch
	fmt.Println("unblock reading", id)

	close(ch)
	q.lock.Lock()
	delete(q.msgs, id)
	q.lock.Unlock()

	return response
}

func (q *queue) Signal(r response) {
	ch := q.getChan(r.ID)
	fmt.Println("blocking sending", r.ID)
	ch <- r
	fmt.Println("unblocked sending", r.ID)
}
