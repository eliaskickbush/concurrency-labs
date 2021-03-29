package main

import (
	"fmt"
	"time"
)

func or(channels ...<-chan interface{}) <-chan interface{} {

	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})

	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[2]:
			case <-channels[3]:
			case <-or(append(channels[3:], orDone)...):
			}
		}

	}()

	return orDone
}

func AfterInterfaceNil(duration time.Duration) <-chan interface{} {

	resStream := make(chan interface{})

	go func() {
		defer close(resStream)
		after := time.After(duration)
		<-after
		fmt.Println(duration)
		resStream <- 1
	}()

	return resStream
}

func main() {

	<-or(AfterInterfaceNil(time.Second), AfterInterfaceNil(time.Millisecond))
}
