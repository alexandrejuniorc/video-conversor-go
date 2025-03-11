package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"time"

	"github.com/alexandrejuniorc/video-conversor-go/pkg/workerpool"
)

type JobNumber struct {
	Number int
}
type ResultNumber struct {
	Value     int
	WorkerID  int
	Timestamp time.Time
}

func processNumber(ctx context.Context, job workerpool.Job) workerpool.Result {
	number := job.(JobNumber).Number
	workerID := number % 3 // just to simulate a worker id

	sleepTime := time.Duration(800+rand.Intn(400)) * time.Millisecond
	time.Sleep(sleepTime)

	return ResultNumber{
		Value:     number,
		WorkerID:  workerID,
		Timestamp: time.Now(),
	}
}

func main() {
	maxValue := 20
	bufferSize := 10

	pool := workerpool.New(processNumber, workerpool.Config{
		WorkerCount: 3,
	})

	inputCh := make(chan workerpool.Job, bufferSize)
	ctx := context.Background()

	resultCh, err := pool.Start(ctx, inputCh)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(maxValue)

	fmt.Println("Starting worker pool with", maxValue, "numbers")

	go func() {
		for i := 0; i < maxValue; i++ {
			inputCh <- JobNumber{Number: i}
		}
		close(inputCh)
	}()

	go func() {
		for result := range resultCh {
			r := result.(ResultNumber)
			fmt.Printf("Number: %d, WorkerID: %d, Timestamp: %s\n", r.Value, r.WorkerID, r.Timestamp.Format(time.RFC3339))
			wg.Done()
		}
	}()

	wg.Wait()
	fmt.Printf("worker pools finished processing %d numbers\n", maxValue)
}
