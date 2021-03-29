package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func FaultyOperation() error {
	time.Sleep(time.Duration(rand.Float64()) * 1000 * time.Millisecond)
	if rand.Float64() >= 0.5 {
		return errors.New("faulty operation went wrong")
	}
	return nil
}

type Result struct {
	Error error
}

func ProcessData(urls []string) <-chan Result {
	resultStream := make(chan Result, len(urls))

	go func() {
		var wg = sync.WaitGroup{}
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				resultStream <- Result{Error: FaultyOperation()}
			}(url)
		}
		wg.Wait()
		close(resultStream)
	}()

	return resultStream
}

func main() {
	rand.Seed(time.Now().Unix())

	resStream := ProcessData([]string{"one", "two", "three", "four"})

	errCount := 0
	for r := range resStream {
		if r.Error != nil {
			errCount++
			fmt.Println("We had an error processing an item")

			if errCount > 2 {
				fmt.Println("Too many errors")
			}
		}
	}

}
