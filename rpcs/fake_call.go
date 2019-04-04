package rpcs

import (
	"errors"
	"github.com/ngaut/log"
	"raft/config"
	"sync"
)

type FakeCall struct {
	Ch      chan struct{}
	mu      sync.RWMutex
	counter int
}

func NewFakeCall() *FakeCall {
	return &FakeCall{make(chan struct{}, 1), sync.RWMutex{}, 0}
}

func (c *FakeCall) Call(name, server string, args, reply interface{}) error {
	log.Debugf("faleCall name: %v server: %v", name, server)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter++
	if c.counter >= len(config.Config.Servers) {
		c.Ch <- struct{}{}
		return errors.New("count limit reached")
	}
	return nil
}
