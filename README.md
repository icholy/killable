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
  err := k.Err()
  fmt.Println("Dying because: ", err)
}()

killable.Defer(k, func() {
  fmt.Println("Dead")
})

killable.Go(func() error {
  time.Sleep(5 * time.Second)
  fmt.Println("Finished sleeping, i'll be dead soon")
  return nil
})

k.Kill(fmt.Errorf("it's time to die!"))
```

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

See `examples/` directory to see how to use it.

The methods like `Defer`, `Go`, `Do`, etc ...  have been placed in the packages because the `Killable` type is meant to be embedded. The interface the `Killable` type exposes makes sense without understanding the `killable` package.

