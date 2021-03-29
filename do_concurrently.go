package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func main() {
	done := make(chan bool)
	httpClient := http.Client{
		Timeout: time.Millisecond * 500,
	}
	go ConcurrentRequests(httpClient, done)
	<-done
}

func ConcurrentRequests(client http.Client, done chan bool) {

	// api := "https://jsonplaceholder.typicode.com"

	// httpClient := http.Client{}
	// Fetch N users
	n := 50
	urls := []string{}
	for i := 0; i < n; i++ {
		urls = append(urls, fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", (i+1)))
	}

	urls = append(urls, "invalid")
	urls = append(urls, "https://valid-but-nonexistant")

	nameStream := FetchURLs(client, urls)

	for el := range nameStream {
		fmt.Println(el)
	}

	done <- true
}

func FetchURLs(client http.Client, urls []string) <-chan string {
	nameStream := make(chan string, len(urls))
	go func() {
		wg := sync.WaitGroup{}
		for _, url := range urls {
			wg.Add(1)
			go func(url string, nameStream chan<- string) {
				defer wg.Done()

				r, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					fmt.Println("error creating request: " + err.Error())
					return
				}

				resp, err := client.Do(r)
				if err != nil {
					fmt.Println("error doing request: " + err.Error())
					return
				}

				if resp.StatusCode != http.StatusOK {
					fmt.Printf("unexpected status code: %d\n", resp.StatusCode)
					return
				}

				unpackingMap := make(map[string]interface{})
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("error reading body: " + err.Error())
					return
				}
				err = json.Unmarshal(bodyBytes, &unpackingMap)
				if err != nil {
					fmt.Println("error unmarshalling: " + err.Error())
					return
				}

				if name, exists := unpackingMap["name"]; !exists {
					fmt.Println("name not found")
					return
				} else {
					if nameString, ok := name.(string); ok {
						nameStream <- nameString
					} else {
						fmt.Println("error with name type")
						return
					}
				}
			}(url, nameStream)
		}
		wg.Wait()
		close(nameStream)
	}()
	return nameStream
}
