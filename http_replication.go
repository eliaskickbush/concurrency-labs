package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func orDone(done <-chan struct{}, values <-chan interface{}) <-chan interface{} {
	resultStream := make(chan interface{})

	go func() {
		defer close(resultStream)
		loop:
		for {
			select {
				case <-done:
					break loop
				case value,open := <- values:
					if !open {
						break loop
					}
					select {
						case <-done:
						case resultStream <- value:
					}
			}
		}

	}()
	return values
}

func fanIn(done <-chan struct{}, channels ...<-chan interface{}) <-chan interface{} {
	resultStream := make(chan interface{})

	multiplex := func(done <-chan struct{}, channel <-chan interface{}, group *sync.WaitGroup){
		defer group.Done()
		mixin := orDone(done, channel)
		for value := range mixin {
			select {
				case <-done:
				case resultStream <- value:
			}
		}
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(channels))
	for _,channel := range channels{
		go multiplex(done, channel, &waitGroup)
	}

	go func() {
		waitGroup.Wait()
		close(resultStream)
	}()
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

type Result struct {
	ID int
	Error error
	Response *http.Response
}

func Replicate(ctx context.Context, request *http.Request, timeout time.Duration, factor int) (*http.Response,error){
	client := http.Client{
		Timeout: timeout,
	}
	resultStream := make(chan Result)

	doRequest := func(id int, ctx context.Context, request *http.Request, group *sync.WaitGroup){
		defer group.Done()
		begin := time.Now()
		req := request.WithContext(ctx)
		resp,err := client.Do(req)
		if err != nil || resp.StatusCode == http.StatusOK{
			select {
				case resultStream <- Result{
					ID: id,
					Error: err,
				}:
				case <-ctx.Done():
					fmt.Println("Preempted #", id, " before reporting error due to context cancelling")
					return
			}
		}
		select {
			case resultStream <- Result{
				ID: id,
				Response: resp,
			}:
				fmt.Println("Finished #", id, " in ", time.Now().Sub(begin).String())
			case <-ctx.Done():
				fmt.Println("Preempted #", id, " before returning request due to context cancelling")
		}
	}

	var wg sync.WaitGroup
	commonContext,cancelFurtherRequests := context.WithCancel(ctx)
	defer cancelFurtherRequests()
	wg.Add(factor)
	for id := 0; id < factor; id++ {
		go doRequest(id, commonContext, request, &wg)
	}

	go func() {
		wg.Wait()
		close(resultStream)
	}()

	for value := range resultStream{
		if value.Error == nil {
			return value.Response, nil
		}
	}



	return nil, errors.New("no requests succeedeed")
}

func main(){
	ctx := context.Background()
	req, _ := http.NewRequest(http.MethodGet, "https://jsonplaceholder.typicode.com/users/1", nil)
	resp, err := Replicate(ctx, req, time.Second, 10)
	if err != nil {
		fmt.Println("got an error from replicator:", err.Error())
	}else{
		fmt.Println("Got a response from replicator:")
		body,err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(body))
		err = resp.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}