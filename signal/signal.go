package signal

import (
	"github.com/ngaut/log"
	"math/rand"
	"time"
)

type observer interface {
	update()
}

type subject interface {
	attach(ob observer)
	notify()
}

type heartBeatSub struct {
	observers []observer
}

func newHeartBeat() *heartBeatSub {
	return &heartBeatSub{[]observer{}}
}

func (s *heartBeatSub) Beat() {
	s.notify()
}

func (s *heartBeatSub) attach(ob observer) {
	s.observers = append(s.observers, ob)
}

func (s heartBeatSub) notify() {
	for _, o := range s.observers {
		o.update()
	}
}

type heartBeatObserver struct {
	t *timeoutSubject
	c *int
}

func newHeartBeatObserver(s *timeoutSubject) *heartBeatObserver {
	return &heartBeatObserver{t: s, c: new(int)}
}

func (h heartBeatObserver) update() {
	log.Debug("heartBeatObserver update")
	h.t.resetTimer()
	*h.c++
}

type timeoutSubject struct {
	observers []observer

	timer *time.Timer
}

func (t *timeoutSubject) resetTimer() {
	ri := rand.Intn(10) + 5
	t.timer.Reset(time.Second * time.Duration(ri))
}

func newTimeoutSub() *timeoutSubject {
	return &timeoutSubject{observers: []observer{}, timer: time.NewTimer(time.Second)}
}

func (t *timeoutSubject) randomTimeoutForever() {
	go func() {
		for {
			select {
			case <-t.timer.C:
				t.resetTimer()
				t.notify()
			}
		}
	}()
}

func (t *timeoutSubject) attach(ob observer) {
	t.observers = append(t.observers, ob)
}

func (t *timeoutSubject) notify() {
	for _, o := range t.observers {
		o.update()
	}
}

type timeoutObserver struct {
	timer *time.Timer
}

func newTimeoutObserver() *timeoutObserver {
	return &timeoutObserver{
		timer: time.NewTimer(time.Duration(time.Second * 5)),
	}
}

func (c timeoutObserver) update() {
	log.Debug("timeoutObserverr update")
}
