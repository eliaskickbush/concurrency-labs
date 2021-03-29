package main

import (
	"fmt"
	"math/rand"
	"time"
)

func ConstantGenerator(constant int, done chan bool) <-chan int {
	constantChannel := make(chan int)

	go func() {
		defer close(constantChannel)
		for {
			select {
			case constantChannel <- constant:
			case <-done: // Clean up logic
				done <- true
				fmt.Println("Cleanup logic")
				return
			}
		}
	}()

	return constantChannel
}

func RandomGenerator(done chan bool) <-chan float64 {
	randStream := make(chan float64)

	go func() {
		defer close(randStream)

		for {
			select {
			case randStream <- rand.Float64():
			case <-done:
				fmt.Println("Clearing logic for Random Generator")
				done <- true
				return
			}
		}

	}()

	return randStream
}

func main() {
	rand.Seed(time.Now().Unix())
	done := make(chan bool)

	randomStream := RandomGenerator(done)
	for i := 0; i < 3; i++ {
		fmt.Printf("%f\n", <-randomStream)
	}
	done <- true
	<-done

	constantStream := ConstantGenerator(5, done)
	for i := 0; i < 3; i++ {
		fmt.Printf("%d\n", <-constantStream)
	}
	done <- true
	<-done
}
