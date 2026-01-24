package image

import (
	"runtime"
	"sync"
)

// WorkerPool manages parallel image processing using CPU cores
type WorkerPool struct {
	workers int
	tasks   chan func()
	wg      sync.WaitGroup
}

// NewWorkerPool creates a new worker pool with NumCPU workers
func NewWorkerPool() *WorkerPool {
	workers := runtime.NumCPU()
	pool := &WorkerPool{
		workers: workers,
		tasks:   make(chan func(), workers*2),
	}

	// Start workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for task := range p.tasks {
		task()
	}
}

// Submit submits a task to the worker pool
func (p *WorkerPool) Submit(task func()) {
	p.tasks <- task
}

// Close closes the worker pool
func (p *WorkerPool) Close() {
	close(p.tasks)
	p.wg.Wait()
}


