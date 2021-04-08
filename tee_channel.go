package main

import (
	"bufio"
	"fmt"
	"os"
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

func tee(done <-chan struct{}, in <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})

	go func() {
		defer close(out1)
		defer close(out2)

		for v := range orDone(done, in) {
			var out1, out2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
				case out1 <- v:
					out1 = nil
				case out2 <- v:
					out2 = nil
				}
			}
		}

	}()

	return out1, out2
}

func Printer(id string, done <-chan struct{}, values <-chan interface{}) {
	for v := range orDone(done, values) {
		fmt.Printf("id(%s): %v\n", id, v)
	}
}

func main() {

	ch := make(chan interface{})
	done := make(chan struct{})

	ch1, ch2 := tee(done, ch)
	go Printer("ch1", done, ch1)
	go Printer("ch2", done, ch2)

	ch <- "tuvieja"

	close(ch)
	rdr := bufio.NewReader(os.Stdin)
	rdr.ReadString('\n')

}
