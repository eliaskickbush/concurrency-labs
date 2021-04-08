package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Operation based heartbeat
func GenerateNumbers(done <-chan struct{}, numbers ...int) (<-chan struct{}, <-chan interface{}) {
	heartbeat := make(chan struct{}, 1)
	stream := make(chan interface{})

	go func() {
		defer close(heartbeat)
		defer close(stream)

		for _, number := range numbers {
			fmt.Println("About to process: ", number)
			select {
			case heartbeat <- struct{}{}:
			default:
			}
			select {
			case <-done:
				break
			case stream <- number:
				break
			}
		}
	}()

	return heartbeat, stream
}

func OperationHeartbeats() {
	done := make(chan struct{})
	heartbeatStream, numberStream := GenerateNumbers(done, 1, 2, 3, 4, 5)

outer:
	for {
		select {
		case _, open := <-heartbeatStream:
			if open {
				fmt.Println("pulse")
			}
		case number, open := <-numberStream:
			if !open {
				break outer
			}
			fmt.Println("Number: ", number)
		case <-time.After(time.Millisecond):
			fmt.Println("Ending prematurely due to delays in heartbeats")
			close(done)
		}
	}
}

func PerformHeavyComputation(done <-chan struct{}, pulseInterval time.Duration) (<-chan struct{}, <-chan interface{}) {
	resultStream := make(chan interface{})
	heartbeat := make(chan struct{})

	go func() {
		defer close(resultStream)

		t := time.NewTicker(pulseInterval)
		expensiveOperation := time.After(time.Duration(rand.Intn(1000))*time.Millisecond + 3*time.Second)
		sentHeartbeat := false
		for {
			select {
			case <-t.C:
				if !sentHeartbeat {
					sentHeartbeat = true
					select {
					case heartbeat <- struct{}{}:
					default:
					}
				}
			case <-done:
				return
			case <-expensiveOperation:
				for {
					select {
					case <-done:
						return
					case resultStream <- "result":
					case <-t.C:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					}
				}
			}
		}
	}()

	return heartbeat, resultStream
}

func main() {
	done := make(chan struct{})
	timeout := time.Second * 2

	heartbeat, resultStream := PerformHeavyComputation(done, time.Millisecond*200)

loop:
	for {
		select {
		case <-heartbeat:
			fmt.Println("pulse")
		case <-time.After(timeout):
			fmt.Println("Preempting due to hearbeat failure")
			break loop
		case <-resultStream:
			fmt.Println("Got result")
			break loop
		}
	}
	close(done)

}
