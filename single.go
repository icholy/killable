package killable

import "sync"

type single struct {
	errc   chan error
	dyingc chan struct{}
	deadc  chan struct{}
	err    error
	wg     sync.WaitGroup
}

func (c *single) add()                   { c.wg.Add(1) }
func (c *single) done()                  { c.wg.Done() }
func (c *single) wait()                  { c.wg.Wait() }
func (c *single) Dying() <-chan struct{} { return c.dyingc }
func (c *single) Dead() <-chan struct{}  { return c.deadc }

func (c *single) errorHandler() {
	c.err = <-c.errc
	close(c.dyingc)
	c.wait()
	close(c.deadc)
}

func (c *single) Kill(reason error) {
	select {
	case c.errc <- reason:
	case <-c.dyingc:
	}
}

func (c *single) Err() error {
	<-c.dyingc
	return c.err
}

func New() Killable {
	c := &single{
		dyingc: make(chan struct{}),
		deadc:  make(chan struct{}),
		errc:   make(chan error),
	}
	go c.errorHandler()
	return c
}
