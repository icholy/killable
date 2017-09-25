package killable

import (
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

func doneTimeout(done <-chan struct{}) bool {
	runtime.Gosched()
	select {
	case <-done:
		return false
	case <-time.After(time.Millisecond * 500):
		return true
	}
}

func doneTimeoutErr(t *testing.T, done <-chan struct{}, msg string) {
	if doneTimeout(done) {
		t.Errorf(msg)
	}
}

func TestKillShouldCloseDyingAndDead(t *testing.T) {
	k := New()
	k.Kill(nil)
	doneTimeoutErr(t, k.Dying(), "Dying chan not closed")
	doneTimeoutErr(t, k.Dead(), "Dying chan not closed")
}

func TestErrShouldBlock(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})
		err  = errors.New("oohh noooess")
	)
	go func() {
		if k.Err() != err {
			t.Errorf("Err() returned wrong error")
		}
		close(done)
	}()
	k.Kill(err)
	doneTimeoutErr(t, done, "Err() didn't return")
}

func TestDoShouldReturnErr(t *testing.T) {
	var (
		k   = New()
		err = errors.New("the error")
	)
	if Do(k, func() error {
		return err
	}) != err {
		t.Errorf("Do didn't return correct error")
	}
}

func TestDoShouldReturnErrDying(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})
	)
	go func() {
		if Do(k, func() error {
			time.Sleep(time.Second)
			return nil
		}) != ErrDying {
			t.Errorf("Do didn't return ErrDying")
		}
		close(done)
	}()
	k.Kill(nil)
	doneTimeoutErr(t, done, "Do didn't return")
}

func TestDoShouldNotRunWhenDying(t *testing.T) {
	k := New()
	k.Kill(nil)
	if Do(k, func() error {
		t.Errorf("Do should not run if Killable is Dead")
		return nil
	}) != ErrDying {
		t.Errorf("Do should return ErrDying if invoked on a Dying Killable")
	}
}

func TestShouldNotDieUntilDoReturns(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})
	)
	Do(k, func() error {
		defer close(done)
		k.Kill(nil)
		<-k.Dying()
		if !doneTimeout(k.Dead()) {
			t.Errorf("Killable should not be dead until Do returns")
		}
		return nil
	})
	<-done
	doneTimeoutErr(t, k.Dead(), "Killable should be dead after Do returns")
}

func TestShouldNotDieUntilMultipleDoReturns(t *testing.T) {
	var (
		k  = New()
		wg sync.WaitGroup
	)
	fn := func() error {
		wg.Done()
		defer wg.Done()
		<-k.Dying()
		if !doneTimeout(k.Dead()) {
			t.Errorf("Killable should not be dead until Do returns")
		}
		return nil
	}
	wg.Add(4)
	go Do(k, fn)
	go Do(k, fn)
	go Do(k, fn)
	go Do(k, fn)
	wg.Wait()
	wg.Add(4)
	k.Kill(nil)
	wg.Wait()
	doneTimeoutErr(t, k.Dead(), "Killable should be dead after Do returns")
}

func TestGoShouldKillWithReturnedError(t *testing.T) {
	var (
		k   = New()
		err = errors.New("this is my error, there are many like it, but this one is mine")
	)
	Go(k, func() error {
		return err
	})
	doneTimeoutErr(t, k.Dying(), "Killable didn't enter dying state")
	if k.Err() != err {
		t.Errorf("Killable didn't recieve correct error")
	}
}

func TestGoShouldNotKillWithReturnedNil(t *testing.T) {
	k := New()
	Go(k, func() error {
		return nil
	})
	if !doneTimeout(k.Dying()) {
		t.Errorf("Go should not have killed Killable")
	}
}

func TestGoShouldNotRunWhenDying(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})
	)
	k.Kill(nil)
	<-k.Dying()
	Go(k, func() error {
		close(done)
		return nil
	})
	if !doneTimeout(done) {
		t.Errorf("Go executed while in dying state")
	}
}

func TestShouldNotDieUntilGoCompletes(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})
	)
	Go(k, func() error {
		defer close(done)
		k.Kill(nil)
		<-k.Dying()
		if !doneTimeout(k.Dead()) {
			t.Errorf("Killable should not be dead until Go completes")
		}
		return nil
	})
	<-done
	doneTimeoutErr(t, k.Dead(), "Killable should be dead after Go completes")
}

func TestShouldNotDieUntilMultipleGoComplete(t *testing.T) {
	var (
		k  = New()
		wg sync.WaitGroup
	)
	fn := func() error {
		defer wg.Done()
		k.Kill(nil)
		<-k.Dying()
		if !doneTimeout(k.Dead()) {
			t.Errorf("Killable should not be dead until Go completes")
		}
		return nil
	}
	wg.Add(4)
	Go(k, fn)
	Go(k, fn)
	Go(k, fn)
	Go(k, fn)
	wg.Wait()
	doneTimeoutErr(t, k.Dead(), "Killable should be dead after Go completes")
}

func TestDeferShouldRunWhenDead(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})
	)
	Defer(k, func() {
		close(done)
	})
	k.Kill(nil)
	doneTimeoutErr(t, done, "Defer callback was not invoked")
}

func TestDeferExecutesOppositeOrder(t *testing.T) {
	var (
		k    = New()
		done = make(chan struct{})

		secondWasExecuted bool
	)
	Defer(k, func() {
		if !secondWasExecuted {
			t.Fatal("Second Defer func should have executed first")
		}
		close(done)
	})
	Defer(k, func() {
		secondWasExecuted = true
	})
	k.Kill(nil)
	doneTimeoutErr(t, done, "Defer callback was never called")
}

func TestKillGroupKillsChildren(t *testing.T) {
	var (
		k1  = New()
		k2  = New()
		g   = New(k1, k2)
		err = errors.New("blah blah blah")
	)
	g.Kill(err)
	doneTimeoutErr(t, k1.Dying(), "killable 1 didn't die")
	if k1.Err() != err {
		t.Errorf("k1 didn't get killed with correct error")
	}
	doneTimeoutErr(t, k2.Dying(), "killable 2 didn't die")
	if k2.Err() != err {
		t.Errorf("k2 didn't get killed with correct error")
	}
}

func TestKillChildKillsGroup(t *testing.T) {
	var (
		k1  = New()
		k2  = New()
		g   = New(k1, k2)
		err = errors.New("foo bar")
	)
	k2.Kill(err)
	doneTimeoutErr(t, k1.Dying(), "killable 1 didn't die")
	if k1.Err() != err {
		t.Errorf("k1 didn't get killed with correct error")
	}
	doneTimeoutErr(t, g.Dying(), "group didn't die")
	if g.Err() != err {
		t.Errorf("group didn't get killed with correct error")
	}
}

func TestKillLocalChildDoesNotKillGroup(t *testing.T) {
	var (
		k = New()
		g = New(k)
	)
	k.Kill(ErrKillLocal)
	if !doneTimeout(g.Dying()) {
		t.Fatal("group got killed by local child error")
	}
}

func TestGroupDeadAfterChildrenComplete(t *testing.T) {
	var (
		k1 = New()
		k2 = New()
		g  = New(k1, k2)
		wg sync.WaitGroup
	)
	fn := func() error {
		defer wg.Done()
		if !doneTimeout(g.Dead()) {
			t.Errorf("group didn't wait for child 1 to complete")
		}
		return nil
	}
	wg.Add(2)
	Go(k1, fn)
	Go(k2, fn)
	g.Kill(nil)
	wg.Wait()
	doneTimeoutErr(t, g.Dead(), "group didn't die")
}

func TestContextDoneWhenDying(t *testing.T) {
	k := New()
	ctx := k.Context()
	k.Kill(nil)
	doneTimeoutErr(t, ctx.Done(), "context isn't done")
}

func TestContextErr(t *testing.T) {
	k := New()
	ctx := k.Context()
	err := errors.New("some error")

	if ctx.Err() != nil {
		t.Fatal("context error should be nil")
	}
	k.Kill(err)
	if ctx.Err() != err {
		t.Fatalf("%s err doesn't equal %s", ctx.Err(), err)
	}
}
