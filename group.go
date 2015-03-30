package killable

import "sync"

type group struct {
	children []Killable
	dyingc   chan struct{}
	deadc    chan struct{}
	errc     chan error
	err      error
	wg       sync.WaitGroup
}

func NewGroup(killables ...Killable) Killable {
	k := &group{
		children: killables,
		dyingc:   make(chan struct{}),
		deadc:    make(chan struct{}),
		errc:     make(chan error),
	}
	k.wg.Add(len(killables))
	for _, child := range killables {
		go k.childErrorHandler(child)
	}
	go k.errorHandler()
	return k
}

func (k *group) add()                   { k.wg.Add(1) }
func (k *group) done()                  { k.wg.Done() }
func (k *group) wait()                  { k.wg.Wait() }
func (k *group) Dying() <-chan struct{} { return k.dyingc }
func (k *group) Dead() <-chan struct{}  { return k.deadc }

func (k *group) childErrorHandler(child Killable) {
	k.Kill(child.Err())
	<-child.Dead()
	k.wg.Done()
}

func (k *group) errorHandler() {
	k.err = <-k.errc
	close(k.dyingc)
	for _, child := range k.children {
		child.Kill(k.err)
	}
	k.wg.Wait()
	close(k.deadc)
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
