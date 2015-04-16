package killable

import (
	"errors"
	"time"
)

var (
	ErrDying      = errors.New("killable: dying")
	ErrKill       = errors.New("killable: killed")
	ErrStillAlive = errors.New("killable: still alive")
)

type Killable interface {
	// Dying is closed immediatly after Kill is called
	Dying() <-chan struct{}

	// Dead is closed after all executing functions have returned
	// These executing functions must have been started with Do or Go
	Dead() <-chan struct{}

	// Put the Killable into the dying state
	Kill(reason error)

	// Return the error passed to Kill
	// blocks until in dying state
	Err() error

	// add a function to run when dead
	addDefer(func())

	// faster state access
	isDead() bool
	isDying() bool

	// access to underlying WaitGroup
	add()
	done()
	wait()
}

func New(killables ...Killable) Killable {
	if len(killables) == 0 {
		return newSingle()
	} else {
		return newGroup(killables...)
	}
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
	if IsDying(k) {
		return ErrDying
	}
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
	if IsDying(k) {
		return
	}
	k.add()
	go func() {
		defer k.done()
		if err := fn(); err != nil && err != ErrDying {
			k.Kill(err)
		}
	}()
}

// Defer invokes a callback once a Killable is dead.
// deferred functions will execute in the opposite order they were defined in
// if the Killable is already dead, the deferred will be called immediatly
func Defer(k Killable, fn func()) { k.addDefer(fn) }

// IsDying returns true if a Killable is in the dying or dead state
func IsDying(k Killable) bool { return k.isDying() }

// IsDead true if the Killable is in the dead state
func IsDead(k Killable) bool { return k.isDead() }

// IsAlive returns true if the Killable isn't in the dead or dying states
func IsAlive(k Killable) bool { return !IsDying(k) }

// Err returns the Killable error
// If the Killable is alive, it returns ErrStillAlive
func Err(k Killable) error {
	if IsDead(k) {
		return k.Err()
	} else {
		return ErrStillAlive
	}
}
