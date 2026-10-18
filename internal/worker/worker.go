package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/NIROOZbx/rag-tutorial.git/internal/document"
)

type Result struct {
	Text string
	Err  error
}

type Job struct {
	Data   []byte
	Result chan Result
}

type Pool struct {
	jobs   chan Job
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

func NewWorkerPool(queueSize, workerCount int) (*Pool, error) {
	if workerCount < 1 {
		return nil, fmt.Errorf("workerCount must be at least 1")
	}
	if queueSize < 0 {
		return nil, fmt.Errorf("queueSize cannot be negative")
	}

	p := Pool{
		jobs: make(chan Job, queueSize),
	}

	p.wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go p.worker()
	}

	return &p, nil
}

func (p *Pool) worker() {

	defer p.wg.Done()

	for job := range p.jobs {
		text, err := document.ExtractPDFText(bytes.NewReader(job.Data))

		job.Result <- Result{
			Text: text,
			Err:  err,
		}
	}

}

func (p *Pool) Submit(ctx context.Context, data []byte) (string, error) {

	resultCh := make(chan Result, 1)

	job := Job{
		Data:   data,
		Result: resultCh,
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return "", errors.New("worker pool is closed")
	}
	select {
	case <-ctx.Done():
		p.mu.Unlock()
		return "", ctx.Err()

	case p.jobs <- job:
		p.mu.Unlock()
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultCh:
		return res.Text, res.Err
	}

}

func (p *Pool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.jobs)
	p.mu.Unlock()
	p.wg.Wait()
}
