package deploy

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const timeout = 5 * time.Minute

// Pool is a bounded worker pool that limits concurrent deploys.
type Pool struct {
	sem    chan struct{}
	wg     sync.WaitGroup
	active atomic.Int64
}

// NewPool creates a pool with max concurrent workers.
// Defaults to 2x CPU cores (reasonable for I/O-bound shell scripts).
func NewPool(maxWorkers int) *Pool {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU() * 2
	}
	return &Pool{
		sem: make(chan struct{}, maxWorkers),
	}
}

// Run submits a script to the worker pool asynchronously.
// If the pool is saturated, the caller blocks until a slot opens.
func (p *Pool) Run(script string) {
	p.wg.Add(1)
	p.sem <- struct{}{} // acquire slot (blocks if full)
	p.active.Add(1)

	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }() // release slot
		defer p.active.Add(-1)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, script)

		output, err := cmd.CombinedOutput()
		out := string(output)
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("deploy timed out: script=%s output=%s", script, out)
			return
		}
		if err != nil {
			log.Printf("deploy failed: script=%s error=%v output=%s", script, err, out)
			return
		}
		log.Printf("deploy succeeded: script=%s output=%s", script, out)
	}()
}

// Active returns the number of currently running deploys.
func (p *Pool) Active() int64 {
	return p.active.Load()
}

// Shutdown waits for all running deploys to finish.
func (p *Pool) Shutdown() {
	p.wg.Wait()
}

// RunBlocking runs a script synchronously with timeout (for testing or single-use).
func RunBlocking(script string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, script)

	output, err := cmd.CombinedOutput()
	out := string(output)
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("deploy timed out")
	}
	if err != nil {
		return out, fmt.Errorf("deploy failed: %w", err)
	}

	return out, nil
}
