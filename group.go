package killable

import (
	"context"
	"sync"
)

type group struct {
	children []Killable
	dyingc   chan struct{}
	deadc    chan struct{}
	errc     chan error
	err      error
	dead     bool
	dying    bool
	deferred []func()
	m        sync.RWMutex
	wg       sync.WaitGroup
}

func newGroup(children ...Killable) Killable {
	k := &group{
		children: children,
		dyingc:   make(chan struct{}),
		deadc:    make(chan struct{}),
		errc:     make(chan error),
	}
	k.wg.Add(len(children))
	for _, child := range children {
		go k.childErrorHandler(child)
	}
	go k.errorHandler()
	return k
}

func (k *group) add()  { k.wg.Add(1) }
func (k *group) done() { k.wg.Done() }
func (k *group) wait() { k.wg.Wait() }

func (k *group) addDefer(fn func()) {
	k.m.Lock()
	k.m.Unlock()
	if k.dead {
		go fn()
	} else {
		k.deferred = append(k.deferred, fn)
	}
}

func (k *group) isDead() bool {
	k.m.RLock()
	defer k.m.RUnlock()
	return k.dead
}

func (k *group) isDying() bool {
	k.m.RLock()
	defer k.m.RUnlock()
	return k.dying
}

func (k *group) Dying() <-chan struct{} { return k.dyingc }
func (k *group) Dead() <-chan struct{}  { return k.deadc }

func (k *group) childErrorHandler(child Killable) {
	if err := child.Err(); err != ErrKillLocal {
		k.Kill(err)
	}
	<-child.Dead()
	k.wg.Done()
}

func (k *group) errorHandler() {
	// wait for an error
	k.err = <-k.errc

	// mark as dying
	close(k.dyingc)
	k.m.Lock()
	k.dying = true
	k.m.Unlock()

	// propagate error to all children
	for _, child := range k.children {
		child.Kill(k.err)
	}

	// wait for all workers to complete
	k.wg.Wait()

	// mark as dead
	close(k.deadc)
	k.m.Lock()
	k.dead = true

	// invoke deferreds
	nDeferred := len(k.deferred)
	for i := range k.deferred {
		k.deferred[nDeferred-i-1]()
	}

	k.m.Unlock()
}

func (k *group) Kill(reason error) {
	select {
	case k.errc <- reason:
	case <-k.dyingc:
	}
}

func (k *group) Err() error {
	<-k.dyingc
	return k.err
}

func (k *group) Context() context.Context {
	return &kContext{k}
}
