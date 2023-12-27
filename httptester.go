package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const timeoutSeconds = 10

// TODO:
// - spread requests randomly within a second
// - add option to repeat x amount of times
// - options for more complex requests (POST with JSON body)
// - record times/statuses and --out to file

type testResult struct {
	status  int
	elapsed time.Duration
}

func (tr testResult) String() string {
	return fmt.Sprintf("{%d %s}", tr.status, tr.elapsed.String())
}

func main() {
	args := os.Args[1:]
	addr := "http://google.com/"
	cnt := 5

	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("Usage: http-tester [url] [count]")
		return
	}

	if len(args) == 2 {
		addr = args[0]

		i, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid argument")
			return
		}
		cnt = i
	}

	makeRequest := func() testResult {
		start := time.Now()
		resp, err := http.Get(addr)
		if err != nil {
			fmt.Println("Error: ", err)
			return testResult{-1, time.Now().Sub(start)}
		}
		fmt.Printf("%s ", resp.Status)

		return testResult{resp.StatusCode, time.Now().Sub(start)}
	}

	result := startTestingLoop(cnt, makeRequest)
	fmt.Printf("Results(%d): %v\n", len(result), result)
}

func startTestingLoop(cnt int, request func() testResult) []testResult {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan testResult, cnt)

	wg.Add(cnt)
	for i := 0; i < cnt; i++ {
		go func() {
			defer wg.Done()
			out <- request()
			select {
			case <-time.After(time.Second * 1):
			case <-ctx.Done():
				return
			}
		}()
	}

	select {
	case <-wait(&wg):
		fmt.Println("Finished")
	case <-time.After(time.Second * timeoutSeconds):
		cancel()
		fmt.Println("Timeout")
	}
	close(out)

	var result []testResult
	for v := range out {
		result = append(result, v)
	}
	return result
}

func wait(wg *sync.WaitGroup) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}
