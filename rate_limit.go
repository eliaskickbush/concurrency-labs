package main

import (
	"bufio"
	"concurrency-labs/ratelimits"
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"os"
	"time"
)

func main() {

	buffer := bufio.NewReader(os.Stdin)
	reading := true
	background := context.Background()
	client := ratelimits.NewConsumerClient(rate.Every(time.Second * 1), 10)

	for reading {
		_,err := buffer.ReadString('\n')
		if err != nil {
			fmt.Println("error reading:", err.Error())
			continue
		}

		ctx, cancel := context.WithTimeout(background, time.Millisecond * 500)
		v,err := client.ExpensiveOperation(ctx)
		if err != nil {
			fmt.Println("Error doing expensive operation: ", err.Error())
		}else{
			fmt.Println(v)
		}
		cancel()
	}


}