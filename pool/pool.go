package pool

import (
	"errors"
	"time"
	"sync/atomic"
	"fmt"
)

const (
	Run int32 = iota
	Stopped
)


// Pool ...
type Pool struct {
	tasks chan func()
	activeWorkers chan struct{}
	stop chan struct{}
	state int32
	numTasks int64
}

// NewPool creates the new Pool instance
func NewPool(size int) *Pool {
	return &Pool{
		tasks: make(chan func()),
		activeWorkers: make(chan struct{}, size),
		stop: make(chan struct{}),
    }
}


// Schedule allows to put the task in the pool for futher processing
func (p *Pool) Schedule(task func()) error {
	if p.isStopped() {
		return errors.New("Pool is stopped.")
	}
	atomic.AddInt64(&p.numTasks, 1)
    select {
	case p.tasks <- task:
	case p.activeWorkers <- struct{}{}:
        go p.process(task)
	}
	return nil
}

// Stop gracefully wait the all outstanding tasks completed with control it takes not more than timeout
func (p *Pool) Stop(timeout time.Duration) error {
	p.setStoppedState()
	close(p.stop)
	timerToStop := time.NewTimer(timeout)
	timerToCheck := time.NewTicker(time.Second)
	defer func() {
		leftTasks := atomic.LoadInt64(&p.numTasks)
		fmt.Printf("Num unprocessed tasks: %v\n", leftTasks)
	}()
	for {
		select {
		case <-timerToStop.C:
			return errors.New("timeout")
		case <-timerToCheck.C:
			if len(p.activeWorkers) == 0 {
				return nil
			}
		}
	}	
}


func (p *Pool) process(task func()) {
	defer func() {
		<- p.activeWorkers
		fmt.Println("Worker is stopped.")
	}()
    for {
		task()
		atomic.AddInt64(&p.numTasks, -1)
		select {
		case <- p.stop:
			return
		case t := <- p.tasks:
			task = t
		}
    }
}

func (p *Pool) isStopped() bool {
	return atomic.LoadInt32(&p.state) == Stopped
}

func (p *Pool) setStoppedState() {
	atomic.StoreInt32(&p.state, Stopped)
}
