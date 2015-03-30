package main

import (
	"fmt"
	"log"
	"time"

	"github.com/icholy/killable"
)

type Worker struct {
	ch chan int64
	killable.Killable
}

func NewWorker() *Worker {
	return &Worker{
		ch:       make(chan int64),
		Killable: killable.New(),
	}
}

func (w *Worker) startProducer() {

	// producer (non-blocking)
	killable.Go(w, func() error {
		defer close(w.ch)
		var i int64

		for {
			select {
			case w.ch <- i:
				i++
			case <-w.Dying():
				return killable.ErrDying
			}

			if i > 100 {
				return fmt.Errorf("limit reached")
			}
		}
		return nil
	})
}

func (w *Worker) consumer() error {
	return killable.Do(w, func() error {
		for i := range w.ch {
			if i == 123 {
				return fmt.Errorf("I don't like 123")
			}
			if err := killable.Sleep(w, 100*time.Millisecond); err != nil {
				return err
			}
			fmt.Printf("got: %d\n", i)
		}
		return nil
	})
}

func (w *Worker) Start() {

	killable.Defer(w, func() {
		fmt.Println("all processes complete, cleaning up")
	})

	killable.Go(w, func() error {
		w.startProducer()
		return w.consumer()
	})
}

func main() {

	w := NewWorker()

	w.Start()

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("Killing the worker")
		w.Kill(nil)
	}()

	if err := w.Err(); err != nil {
		log.Fatal(err)
	}

}
