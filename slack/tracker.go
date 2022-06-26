package slack

import (
	"fmt"
	"sync"
)

type Tracker struct {
	mu       sync.Mutex
	requests map[string]chan bool
}

func NewTracker() *Tracker {
	return &Tracker{
		requests: make(map[string]chan bool),
	}
}

func (t *Tracker) Track(cid string) (chan bool, error) {
	if _, ok := t.requests[cid]; ok {
		return nil, fmt.Errorf("request already exists")
	}
	t.mu.Lock()
	resCh := make(chan bool, 1)
	t.requests[cid] = resCh
	t.mu.Unlock()
	return resCh, nil
}

func (t *Tracker) Update(cid string, status bool) error {
	t.mu.Lock()
	select {
	case t.requests[cid] <- status:
		// default:
		// 	return fmt.Errorf("could not update tracker for %s", cid)
	}
	delete(t.requests, cid)
	t.mu.Unlock()
	return nil
}
