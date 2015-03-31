package killable

import "sync"

type single struct {
	errc   chan error
	dyingc chan struct{}
	deadc  chan struct{}
	err    error
	wg     sync.WaitGroup
}

func (k *single) add()                   { k.wg.Add(1) }
func (k *single) done()                  { k.wg.Done() }
func (k *single) wait()                  { k.wg.Wait() }
func (k *single) Dying() <-chan struct{} { return k.dyingc }
func (k *single) Dead() <-chan struct{}  { return k.deadc }

func (k *single) errorHandler() {
	k.err = <-k.errc
	close(k.dyingc)
	k.wait()
	close(k.deadc)
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
