# Killable (WIP) [![Build Status](https://travis-ci.org/icholy/killable.svg?branch=master)](https://travis-ci.org/icholy/killable)

> A package for graceful shutdowns in Go (inspired by tomb)

## States

A `Killable` represents a group of goroutines. It goes through 3 stages:

0. Alive - The goroutines are running
0. Dying - The goroutines are being signaled to terminate
0. Dead  - All managed goroutines have terminated

![](images/states.jpg)

There are two ways a `Killable` can enter the dying state.

0. One of the managed goroutines returns a non `nil` `error`
0. The `Kill(error)` method is invoked on the `Killable`.

## Managed Goroutines

Goroutines managed by the `Killable` are started with `killable.Go`


``` go
k := killable.New()

go func() {
  <-k.Dying()
  fmt.Println("Dying")

  <-k.Dead()
  fmt.Println("Dead")
}()

killable.Go(k, func() error {
  time.Sleep(5 * time.Second)
  fmt.Println("Finished sleeping, i'll be dead soon")
  return nil
})

k.Kill(fmt.Errorf("it's time to die!"))
```

* A `Killable` is not dead until all managed goroutines have returned.
* If the goroutine returns a non `nil` `error`, the `Killable` starts dying.
* If the `Killable` is already dying when the `Go` method is invoked, it does not run.

## Defer

`Defer` is similar to the `defer` keyword. 

``` go
func Connect(k killable.Killable) (*sql.DB, error) {

  db, err := sql.Open("foo", "bar")
  if err != nil {
    return nil, err
  }

  // clean up resources near instantiation
  // execute in opposite order after Killable is dead
  killable.Defer(k, func() {
    db.Close()
  })

  return db, nil
}
```

* Deferred methods are called once the killable is dead.
* Deferred methods are invoked in the opposite order they were defined (lifo).
* If the `Killable` is already dead, the function is called immediately.

## Linking

`Killable`s can be linked to eachother in a parent/child relationship.

* If a child is killed, the parent is also killed.
* If the parent is killed, it kills all the children.
* If the `reason` is `ErrKillLocal`, the parent ignores it.
* The parent doesn't die until all the children are dead

``` go

func makeChild(d time.Duration) killable.Killable {
  k := killable.New()

  killable.Go(k, func() {
    time.Sleep(d)
    return killable.ErrKill
  })

  return k
}

var (
  // children
  k1 = makeChild(4 * time.Second)
  k2 = makeChild(3 * time.Second)
  k3 = makeChild(2 * time.Second)

  // parent
  k4 = killable.New(k1, k2, k3)
)

killable.Defer(k4, func() {
  fmt.Println("All children are dead!")
})

go func() {
  <-k4.Dying()
  fmt.Println("Killing all children")
}()

```

![](images/killable.gif)

See `examples/` directory.

The methods like `Defer`, `Go`, `Do`, etc ...  have been placed in the packages because the `Killable` type is meant to be embedded. The interface the `Killable` type exposes makes sense without understanding the `killable` package.

