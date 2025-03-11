package workerpool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Job interface{}

type Result interface{}

type ProcessFunc func(ctx context.Context, job Job) Result

type Workerpool interface {
	Start(ctx context.Context, inputCh <-chan Job) (<-chan Result, error)
	Stop() error
	IsRunning() bool
}

type State int

const (
	StateIdle State = iota
	StateRunning
	StateStopping
)

type Config struct {
	WorkerCount int
	Logger      *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		WorkerCount: 1,
		Logger:      slog.Default(),
	}
}

type workerPool struct {
	workerCount int
	processFunc ProcessFunc
	logger      *slog.Logger
	state       State
	stateMutex  sync.Mutex
	stopCh      chan struct{}
	stopWg      sync.WaitGroup
}

func New(processFunc ProcessFunc, config Config) *workerPool {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 1
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	return &workerPool{
		processFunc: processFunc,
		workerCount: config.WorkerCount,
		stopCh:      make(chan struct{}),
		state:       StateIdle,
		logger:      config.Logger,
	}
}

func (wp *workerPool) Start(ctx context.Context, inputCh <-chan Job) (<-chan Result, error) {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock() // vai rodar o unlock no final da função

	if wp.state != StateIdle {
		return nil, fmt.Errorf("workerpool is not idle")
	}

	resultCh := make(chan Result)
	wp.state = StateRunning
	wp.stopCh = make(chan struct{})
	wp.stopWg.Add(wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(ctx, i, inputCh, resultCh)
	}

	go func() {
		wp.stopWg.Wait()
		close(resultCh)

		wp.stateMutex.Lock()
		wp.state = StateIdle
		wp.stateMutex.Unlock()
	}()

	return resultCh, nil
}

func (wp *workerPool) Stop() error {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()

	if wp.state != StateRunning {
		return fmt.Errorf("workerpool is not running")
	}

	wp.state = StateStopping
	close(wp.stopCh)

	wp.stopWg.Wait()
	wp.state = StateIdle

	return nil
}

func (wp *workerPool) IsRunning() bool {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()
	return wp.state == StateRunning
}

func (wp *workerPool) worker(ctx context.Context, id int, inputCh <-chan Job, resultCh chan<- Result) {
	wp.logger.Info("worker started", "worker_id", id)

	for {
		select {
		case <-wp.stopCh:
			wp.logger.Info("worker interrupted", "worker_id", id)
			return
		case <-ctx.Done():
			wp.logger.Info("worker interrupted by context", "worker_id", id)
			return
		case job, ok := <-inputCh:
			if !ok {
				wp.logger.Info("worker interrupted by closed input channel", "worker_id", id)
				return
			}

			result := wp.processFunc(ctx, job)
			select {
			case resultCh <- result:
			case <-wp.stopCh:
				wp.logger.Info("worker interrupted, result not sent", "worker_id", id)
				return
			case <-ctx.Done():
				wp.logger.Info("worker interrupted by context", "worker_id", id)
				return
			}
		}
	}
}
