package main

import (
	"fmt"
)

func numberGenerator(done chan struct{}, numbers ...int) <-chan int {
	numberStream := make(chan int)
	go func() {
		defer close(numberStream)
		for _, number := range numbers {
			select {
			case <-done:
				break
			case numberStream <- number:
			}
		}
	}()
	return numberStream
}

func multiply(done <-chan struct{}, multiplier int, numberStream <-chan int) <-chan int {
	resultStream := make(chan int)

	go func() {
		defer close(resultStream)
		for n := range numberStream {
			select {
			case <-done:
			case resultStream <- (n * multiplier):
			}
		}
	}()

	return resultStream
}

func add(done <-chan struct{}, number int, numberStream <-chan int) <-chan int {
	resultStream := make(chan int)

	go func() {
		defer close(resultStream)
		for n := range numberStream {
			select {
			case <-done:
			case resultStream <- (n + number):
			}
		}
	}()

	return resultStream
}

// TODO: Implement Repeat, Take, Constant and Random generators...

func main() {

	numbers := []int{1, 2, 3, 4, 5}
	done := make(chan struct{})

	numberStream := add(done, 2, multiply(done, 2, numberGenerator(done, numbers...)))

	//rdr := bufio.NewReader(os.Stdin)
	//rdr.ReadString('\n')

	for v := range numberStream {
		fmt.Println(v)
	}

	close(done)

}
