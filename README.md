# Killable

> A packages for graceful shutdowns in Go

A `Killable` has 3 states:

0. Alive - the worker is running
0. Dying - the worker is in the process of shutting down
0. Dead - all worker processes have completed

There are two ways a killable process can terminate.

0. One of the worker functions returns an `error`
0. The `Kill(error)` method is invoked on the `Killable`

Let's look at a simple example of a killable pipeline:

``` go

var (
  k = killable.New()
  c = make(chan int64)
)

// killer
go func() {
  time.Sleep(5 * time.Second)
  k.Kill(nil)
}()

// producer
go func() {

  var i int64
  defer close(out)

  for {
    select {
    case ch <- i:
    case <-k.Dying()
      return
    }
    i++
  }

}()

// consumer
for i := range c {
  fmt.Println(i)
}
```

When `Kill(error)` is called on the `Killable`, it closes the `Dying()` channel.
When recieving on a closed channel, you'll immediatly get a zero value of its type.

Now say the producer wanted to cancel itself:

``` go

var (
  k = killable.New()
  c = make(chan int64)
)

// producer
killable.Go(k, func() error {

  var i int64
  defer close(out)

  for {

    select {
    case ch <- i:
    case <-k.Dying()
      return Killable.ErrDying
    }

    if i > 100 {
      return fmt.Errorf("limit reached")
    }

    i++
  }

  return nil
}()

// consumer
for i := range c {
  fmt.Println(i)
}

if err := k.Err(); err != nil {
  log.Fatal(err)
}
```

In this case, we're using the `killable.Go` function.
It starts a goroutine using the provided function.

* If the function returns an error that is not `nil` or `ErrDying` it will call `Kill` with it.
