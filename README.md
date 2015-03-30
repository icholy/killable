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

See examples directory to see how to use it.
