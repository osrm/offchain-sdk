package worker

import (
	"fmt"

	"github.com/berachain/offchain-sdk/log"
)

// Pool represents a pool of workers.
type Pool interface {
	Start()
	Stop()
	AddTask(Executor)
	RespChan() chan Resultor
}

// pool is a pool of workers.
type pool struct {
	name    string
	logger  log.Logger
	execCh  chan Executor
	resCh   chan Resultor
	workers []*worker
}

// NewPool creates a new pool of workers.
func NewPool(
	name string,
	totalWorkers uint32,
	logger log.Logger,
) Pool {
	// Intialize the pool.
	p := &pool{
		name:    name,
		logger:  logger,
		workers: make([]*worker, 0),
		execCh:  make(chan Executor),
		resCh:   make(chan Resultor),
	}

	// Iterate through the number of workers and create them.
	for i := uint32(0); i < totalWorkers; i++ {
		w := newWorker(
			i,
			p.execCh,
			p.resCh,
			// TODO: don't hardcode stdout.
			logger,
		)
		p.workers = append(p.workers, w)
	}

	return p
}

// Logger returns the logger for the worker.
func (p *pool) Logger() log.Logger {
	return p.logger.With("namespace", fmt.Sprintf("worker-pool-%s", p.name))
}

// Start starts the pool of workers.
func (p *pool) Start() {
	// Start all the workers.
	p.Logger().Info("starting workers")
	for _, w := range p.workers {
		go w.Start()
	}
}

// Stop stops the pool of workers.
func (p *pool) Stop() {
	// Stop all the workers.
	p.Logger().Info("attemping to stop workers")
	for _, w := range p.workers {
		w.Stop()
	}

	// Ensure the channels get closed
	close(p.execCh)
	close(p.resCh)
}

// AddTask adds a task to the pool.
func (p *pool) AddTask(exec Executor) {
	go p.addTask(exec)
}

// addTask adds a task to the pool.
func (p *pool) addTask(exec Executor) {
	p.execCh <- exec
}

// RespChan returns the response channel.
func (p *pool) RespChan() chan Resultor {
	return p.resCh
}

// GetResult gets the most recent result from the pool.
func (p *pool) GetResult() Resultor {
	return <-p.resCh
}