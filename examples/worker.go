package main

import (
	"fmt"
	"log"
	"time"

	"github.com/icholy/killable"
)

type Worker struct {
	name string
	ch   chan int64
	killable.Killable
}

func NewWorker(name string) *Worker {
	return &Worker{
		name:     name,
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
				return fmt.Errorf("worker: %s: limit reached", w.name)
			}
		}
		return nil
	})
}

func (w *Worker) consumer() error {
	return killable.Do(w, func() error {
		for i := range w.ch {
			if i == 123 {
				return fmt.Errorf("worker: %s: I don't like 123", w.name)
			}
			if err := killable.Sleep(w, 100*time.Millisecond); err != nil {
				return err
			}
			fmt.Printf("worker: %s: %d\n", w.name, i)
		}
		return nil
	})
}

func (w *Worker) Start() {

	killable.Defer(w, func() {
		fmt.Printf("worker: %s: all processes complete, cleaning up\n", w.name)
	})

	killable.Go(w, func() error {
		w.startProducer()
		return w.consumer()
	})
}

func main() {

	var (
		w1 = NewWorker("Worker 1")
		w2 = NewWorker("Worker 2")
		w3 = NewWorker("Worker 3")

		g = killable.New(w1, w2, w3)
	)

	killable.Defer(g, func() {
		fmt.Println("All workers are dead")
	})

	w1.Start()
	w2.Start()
	w3.Start()

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("Killing the worker group")
		g.Kill(fmt.Errorf("time to die!"))
	}()

	if err := g.Err(); err != nil {
		log.Println(err)
	}

	// wait for the deferreds to run
	<-g.Dead()
	time.Sleep(time.Second)

}
