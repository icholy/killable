# Killable

> A package for graceful shutdowns in Go

A `Killable` has 3 states:

0. Alive - the worker is running
0. Dying - the worker is in the process of shutting down
0. Dead - all worker processes have completed

There are two ways a killable process can terminate.

0. One of the worker functions returns an `error`
0. The `Kill(error)` method is invoked on the `Killable`

Worker processes can be started with:

* `killable.Go` which starts a goroutine
* `killable.Do` which blocks while executing

``` go
k = killable.New()

// killer
go func() {
  time.Sleep(5 * time.Second)
  k.Kill(nil)
}()

// Defer runs the callback once all worker functions (Go/Do) have returned
killable.Defer(k, func() {
  fmt.Println("all worker function done")
})

killable.Go(k, func () error {

  ch = make(chan int64)

  // producer (non-blocking)
  killable.Go(k, func () error {
    defer close(ch)
    var i int64
    
    for {
      select {
      case ch <- i:
        i++
      case <-k.Dying()
        return killable.ErrDying
      }

      if err := killable.Sleep(k, time.Second); err != nil {
        return err
      }
    
      if i > 100 {
        return fmt.Errorf("limit reached")
      }
    }
    return nil
  })

  // consumer (blocking)
  return killable.Do(k, func() error {
    for i := range ch {
      if i == 123 {
        return fmt.Errorf("I don't like 123")
      }
    }
    return nil
  })

})

if err := k.Err(); err != nil {
  log.Fatal(err)
}

```

Multiple `Killable`s can be joined together in a `Killable` group.
If any of the `Killable`s are killed, all others in the group will be killed as well.

``` go
var (
  k1 = killable.New()
  k2 = killable.New()
  k3 = killable.NewGroup(k1, k2)
)

go func() {
  log.Println("k1", k1.Err())
}()

go func() {
  log.Println("k2", k2.Err())
}()

go func() {
  log.Println("k3", k3.Err())
}()

k2.Kill(fmt.Errorf("time to die"))
```


