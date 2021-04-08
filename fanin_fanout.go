package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func fanIn(done chan struct{}, chans ...<-chan interface{}) <-chan interface{} {
	fanInStream := make(chan interface{})
	wg := &sync.WaitGroup{}
	multiplex := func(ch <-chan interface{}) {
		defer wg.Done()
		for value := range ch {
			select {
			case <-done:
				break
			case fanInStream <- value:
			}
		}
	}

	wg.Add(len(chans))
	for _, channel := range chans {
		go multiplex(channel)
	}

	go func() {
		wg.Wait()
		close(fanInStream)
	}()
	return fanInStream
}

func NilTicker(done <-chan struct{}, d time.Duration) <-chan interface{} {
	resultStream := make(chan interface{})
	ticker := time.NewTicker(d)

	go func() {
		defer close(resultStream)
		for t := range ticker.C {
			select {
			case <-done:
				break
			case resultStream <- t:
			}
		}
	}()

	return resultStream
}

func Range(done <-chan struct{}, n int) <-chan interface{} {
	stream := make(chan interface{})

	go func() {
		defer close(stream)
		for i := 0; i < n; i++ {
			select {
			case <-done:
				break
			case stream <- i:
			}
		}
	}()

	return stream
}

func expensiveComputation(done <-chan struct{}, valueStream <-chan interface{}) <-chan interface{} {
	resultStream := make(chan interface{})
	go func() {
		defer close(resultStream)
		for value := range valueStream {
			// Simulate expensive computation
			delay := time.Duration(1000+rand.Float64()*1000) * time.Millisecond
			fmt.Printf("Sleeping for %d ms\n", delay/time.Millisecond)
			time.Sleep(delay)
			select {
			case <-done:
				break
			case resultStream <- value:
			}
		}
	}()

	return resultStream
}

func main() {
	rand.Seed(time.Now().Unix())
	done := make(chan struct{})
	workers := 5
	computeWorkers := make([]<-chan interface{}, workers)
	numberStream := Range(done, 5)
	for i := 0; i < workers; i++ {
		computeWorkers[i] = expensiveComputation(done, numberStream)
	}
	results := fanIn(done, computeWorkers...)
	for n := range results {
		fmt.Println(n)
	}
}
