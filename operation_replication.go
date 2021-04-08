package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func orDone(done <-chan struct{}, values <-chan interface{}) <-chan interface{}{
	resultStream := make(chan interface{})

	go func() {
		defer close(resultStream)
		loop:
		for {
			select {
				case <-done:
					break loop
				case v,open := <-values:
					if !open {
						break loop
					}
					select {
						case <-done:
						case resultStream <- v:
					}
			}
		}

	}()

	return resultStream
}

func replicated(done <-chan struct{}, factor int) <-chan interface{} {
	resultStream := make(chan interface{})

	go func() {
		defer close(resultStream)
		t := time.After(time.Duration(rand.Intn(3000)))
		for {
			select {
				case <-t:
					select {
						case <-done:
							break
						case resultStream <- 1:
					}
				case <-done:
			}
		}
	}()

	return resultStream
}

func fanIn(done <-chan struct{}, channels ...<-chan interface{}) <-chan interface{}{
	resultStream := make(chan interface{})

	go func() {
		defer close(resultStream)

		var wg sync.WaitGroup

		multiplex := func(
			done <-chan struct{},
			receiveChannel <-chan interface{},
			sendChannel chan<- interface{},
			wg *sync.WaitGroup){
			defer wg.Done()

			receive := orDone(done, receiveChannel)
			for v := range receive {
				select {
					case <-done:
					case sendChannel <- v:
				}
			}
		}
		wg.Add(len(channels))
		for _,channel := range channels {
			go multiplex(done, channel, resultStream, &wg)
		}
		wg.Wait()
	}()

	return resultStream
}

func AfterInterface(duration time.Duration) <-chan interface{}{
	resultStream := make(chan interface{})

	time.AfterFunc(duration, func() {
		resultStream <- 1
		close(resultStream)
	})

	return resultStream
}

func TickerInterface(done <-chan struct{}, interval time.Duration) <-chan interface{}{
	resultStream := make(chan interface{})
	go func() {
		defer close(resultStream)
		ticker := time.NewTicker(interval)
		loop:
		for {
			outValue := <-ticker.C
			select {
				case <-done:
					break loop
				case resultStream <- outValue:
				default:
			}
		}
	}()
	return resultStream
}

func Replicate(done chan struct{}, factor int) (interface{},chan struct{}){
	resultStream := make(chan interface{}, factor)
	ready := make(chan struct{})

	var wg sync.WaitGroup
	expensiveOperation := func(id int, wg *sync.WaitGroup,resultStream chan<- interface{}){
		defer wg.Done()
		before := time.Now()
		c := time.After(time.Duration(rand.Intn(1000)) * time.Millisecond)
		v := <-c
		fmt.Println("id: ", id, " ", v.Sub(before).String())
		resultStream <- id
	}

	wg.Add(factor)
	for i := 0; i < factor; i++ {
		go expensiveOperation(i, &wg, resultStream)
	}

	go func() {
		wg.Wait()
		close(resultStream)
		ready <- struct{}{}
	}()

	firstResult := <-resultStream
	return firstResult, ready
}

func main() {
	rand.Seed(time.Now().Unix())

	done := make(chan struct{})

	result, ready := Replicate(done, 100)
	fmt.Println(result)
	<-ready

}
