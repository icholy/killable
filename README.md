# Killable (WIP) 

> A package for graceful shutdowns in Go

**Note:** The API is still in flux.

A `Killable` represents a group of goroutines. It goes through 3 stages:

0. Alive - The goroutines are running
0. Dying - The goroutines are being signaled to terminate
0. Dead  - All goroutines have terminated

There are two ways a `Killable` can enter the dying state.

0. One of the goroutines returns an `error`
0. The `Kill(error)` method is invoked on the `Killable`

Goroutines managed by the `Killable` can be started with:

* `killable.Go` which starts a goroutine.
* `killable.Do` which blocks while executing.

``` go
k := killable.New()

go func() {
  <-k.Dying()
  fmt.Println("Dying")
}()

go func() {
  <-k.Dead()
  fmt.Println("Dead")
}()

// create managed goroutine
killable.Go(k, func() error {
  time.Sleep(5*time.Second)
  fmt.Println("Finished Sleeping")
  return nil
})

k.Kill(nil)
```

The `Err()` and `Defer()` methods make it easier to make use of the states.

``` go
k := killable.New()

// Err will block until the Killable is in the Dying state
go func() {
  err := k.Err()
  fmt.Println("Dying: ", err)
}()

// Defer will execute the function after the Killable is dead
killable.Defer(k, func() {
  fmt.Println("Dead")
})

// create managed goroutine
killable.Go(k, func() error {
  time.Sleep(5*time.Second)

  // returning a non nil error will put
  // the killable in a dying state
  return killable.ErrKill
})
```

See `examples/` directory to see how to use it.

The methods like `Defer`, `Go`, `Do`, etc ...  have been placed in the packages because the `Killable` type is meant to be embedded. The interface the `Killable` type exposes makes sense without understanding the `killable` package.

