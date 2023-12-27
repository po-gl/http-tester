package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
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
	cnt := flag.Int("n", 5, "count of test requests to make")
	reps := flag.Int("reps", 1, "repetitions of the test to make")
	nospread := flag.Bool("no-spread", false, "disable randomly spreading requests within one second per repetition")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Arguments:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  URL   the destination URL to test\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.Arg(0) == "" {
		flag.Usage()
		return
	}
	addr := flag.Arg(0)

	makeRequest := func() testResult {
		t := time.Now()
		resp, err := http.Get(addr)
		if err != nil {
			fmt.Println("Error: ", err)
			return testResult{-1, time.Now().Sub(t)}
		}
		fmt.Printf("%s ", resp.Status)

		return testResult{resp.StatusCode, time.Now().Sub(t)}
	}

	for i := 0; i < *reps; i++ {
		t := time.Now()
		result := startTestingLoop(*cnt, *nospread, makeRequest)
		fmt.Printf("%d Results(%d) in %s: %v\n", i+1, len(result), time.Now().Sub(t).String(), result)
	}
}

func startTestingLoop(cnt int, nospread bool, request func() testResult) []testResult {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan testResult, cnt)

	wg.Add(cnt)
	for i := 0; i < cnt; i++ {
		go func() {
			defer wg.Done()
			if !nospread {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			}
			select {
			case out <- request():
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
