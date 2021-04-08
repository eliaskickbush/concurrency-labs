package main

import (
	"fmt"
	"time"
)

func orDone(done <-chan struct{}, values <-chan interface{}) <-chan interface{} {
	resultStream := make(chan interface{})

	go func() {
		defer close(resultStream)
		for {
			select {
			case <-done:
				break
			case value, open := <-values:
				if !open {
					return
				}
				select {
				case resultStream <- value:
				case <-done:
				}
			}
		}

	}()

	return resultStream
}

func Debounce(done <-chan struct{}, valueStream <-chan interface{}, interval time.Duration) <-chan interface{} {
	resultStream := make(chan interface{})
	ticker := time.NewTicker(interval)

	go func() {
		defer close(resultStream)
		for value := range valueStream {
			select {
			case <-done:
				return
			case <-ticker.C:
				select {
				case resultStream <- value:
				case <-done:
				}
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
				return
			case stream <- i:
			}
		}
	}()

	return stream
}

func main() {

	done := make(chan struct{})

	time.AfterFunc(time.Second*2, func() {
		close(done)
	})

	for v := range orDone(done, Debounce(done, Range(done, 1000000), time.Millisecond)) {
		fmt.Println(v)
	}

}
