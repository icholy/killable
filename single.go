package killable

import "sync"

type single struct {
	errc   chan error
	dyingc chan struct{}
	deadc  chan struct{}
	dead   bool
	dying  bool
	err    error
	wg     sync.WaitGroup
	m      sync.RWMutex
}

func (k *single) add()  { k.wg.Add(1) }
func (k *single) done() { k.wg.Done() }
func (k *single) wait() { k.wg.Wait() }

func (k *single) isDying() bool {
	k.m.RLock()
	defer k.m.RUnlock()
	return k.dying
}

func (k *single) isDead() bool {
	k.m.RLock()
	defer k.m.RUnlock()
	return k.dead
}

func (k *single) Dying() <-chan struct{} { return k.dyingc }
func (k *single) Dead() <-chan struct{}  { return k.deadc }

func (k *single) errorHandler() {
	// wait for error
	k.err = <-k.errc

	// mark as dying
	close(k.dyingc)
	k.m.Lock()
	k.dying = true
	k.m.Unlock()

	// wait for workers to complete
	k.wait()

	// mark as dead
	close(k.deadc)
	k.m.Lock()
	k.dead = true
	k.m.Unlock()
}

func (k *single) Kill(reason error) {
	select {
	case k.errc <- reason:
	case <-k.dyingc:
	}
}

func (k *single) Err() error {
	<-k.dyingc
	return k.err
}

func newSingle() Killable {
	k := &single{
		dyingc: make(chan struct{}),
		deadc:  make(chan struct{}),
		errc:   make(chan error),
	}
	go k.errorHandler()
	return k
}
