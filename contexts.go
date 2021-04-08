package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	resChan := ContextAwareProcess(ctx)
	fmt.Println(<-resChan)

}

func ContextAwareProcess(ctx context.Context) <-chan interface{} {
	resultStream := make(chan interface{})

	go func() {
		defer close(resultStream)

		t := time.After(time.Millisecond * 200)

		fmt.Println(string(debug.Stack()))

		select {
		case <-ctx.Done():
			fmt.Println("We got cancelled boys")
			return
		case <-t:
			select {
			case <-ctx.Done():
				fmt.Println("We got cancelled after getting a value boys")
			case resultStream <- 1:
			}
		}
	}()

	return resultStream
}
