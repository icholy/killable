package killable

import (
	"errors"
	"time"
)

var (
	ErrDying      = errors.New("killable: dying")
	ErrStillAlive = errors.New("killable: still alive")
)

type Killable interface {
	// Dying is close immediatly after Kill is called
	Dying() <-chan struct{}

	// Dead is closed after all executing functions have returned
	// These executing functions must have been started with Do or Go
	Dead() <-chan struct{}

	// Put the Killable into the dying state
	Kill(reason error)

	// Return the error passed to Kill
	// blocks until in dying state
	Err() error

	// access to underlying WaitGroup
	add()
	done()
	wait()
}

// Sleep blocks for a specified duration
// If the Killable is marked as dying it will return
// immediatly with ErrDying
func Sleep(k Killable, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-k.Dying():
		return ErrDying
	}
}

// Do executes a function and retuns its error value.
// * If the Killable is marked as dying it will return immediatly with ErrDying.
// * The Killable will not be marked as dead until all calls to Do have returned.
func Do(k Killable, fn func() error) error {
	k.add()
	ch := make(chan error)
	go func() {
		defer k.done()
		select {
		case ch <- fn():
		case <-k.Dying():
		}
	}()
	select {
	case err := <-ch:
		return err
	case <-k.Dying():
		return ErrDying
	}
}

// Go executes a function in a goroutine
// * If the function returns a non-nil error the Killable is killed using that error
// * The Killable will no be marked as dead until all calls to Go have returned
func Go(k Killable, fn func() error) {
	k.add()
	go func() {
		defer k.done()
		if err := fn(); err != nil && err != ErrDying {
			k.Kill(err)
		}
	}()
}

// Defer invokes a callback once a Killable is dead
func Defer(k Killable, fn func()) {
	go func() {
		<-k.Dead()
		fn()
	}()
}

// Dying returns true if a Killable is in the dying or dead state
func Dying(k Killable) bool {
	select {
	case <-k.Dying():
		return true
	default:
		return false
	}
}

// Dead returns true if the Killable is in the dead state
func Dead(k Killable) bool {
	select {
	case <-k.Dead():
		return true
	default:
		return false
	}
}

// Alive returns true if the Killable is not in a dying state
func Alive(k Killable) bool {
	return !Dying(k)
}

// Err returns the Killable error
// If the Killable is alive, it returns ErrStillAlive
func Err(k Killable) error {
	if Alive(k) {
		return ErrStillAlive
	} else {
		return k.Err()
	}
}
