package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const timeoutSeconds = 10

// TODO:
// - refactor into functions/structs
// - commandline arguments
// - spread requests randomly within a second
// - add option to repeat x amount of times
// - options for more complex requests (POST with JSON body)
// - record times/statuses and --out to file

func main() {
	const conc = 10
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan int, conc)

	wg.Add(conc)
	for i := 0; i < conc; i++ {
		go func(i int) {
			defer wg.Done()
			resp, err := http.Get("https://google.com/")
			if err != nil {
				fmt.Println("Http error:", err)
				return
			}
			fmt.Printf("%s ", resp.Status)

			select {
			case <-time.After(time.Second * 1):
			case <-ctx.Done():
				return
			}
			out <- resp.StatusCode
		}(i)
	}

	select {
	case <-wait(&wg):
		fmt.Println("Finished")
	case <-time.After(time.Second * timeoutSeconds):
		cancel()
		fmt.Println("Timeout")
	}
	close(out)

	var result []int
	for v := range out {
		result = append(result, v)
	}
	fmt.Printf("Results(%d): %v\n", len(result), result)
}

func wait(wg *sync.WaitGroup) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}
