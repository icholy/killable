package killable

import (
	"errors"
	"time"
)

var (
	ErrDying      = errors.New("terminator: dying")
	ErrStillAlive = errors.New("terminator: still alive")
)

type Killable interface {
	Dying() <-chan struct{}
	Dead() <-chan struct{}
	Kill(reason error)
	Err() error

	add()
	done()
	wait()
}

func Sleep(k Killable, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-k.Dying():
		return ErrDying
	}
}

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

func Go(k Killable, fn func() error) {
	k.add()
	go func() {
		defer k.done()
		if err := fn(); err != ErrDying {
			k.Kill(err)
		}
	}()
}

func Defer(k Killable, fn func()) {
	go func() {
		<-k.Dead()
		fn()
	}()
}

func Dying(k Killable) bool {
	select {
	case <-k.Dying():
		return true
	default:
		return false
	}
}

func Dead(k Killable) bool {
	select {
	case <-k.Dead():
		return true
	default:
		return false
	}
}

func Alive(k Killable) bool {
	return !Dying(k)
}

func Err(k Killable) error {
	if Alive(k) {
		return ErrStillAlive
	} else {
		return k.Err()
	}
}
