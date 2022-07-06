package xwp

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Handler 处理器
type Handler func(data interface{})

// Worker 工作者
type Worker struct {
	pool    *WorkerPool
	handler Handler
	jobChan JobQueue
	quit    chan bool
}

// NewWorker
func NewWorker(p *WorkerPool) *Worker {
	return &Worker{
		pool: p,
		handler: func(data interface{}) {
			if p.RunF != nil {
				atomic.AddInt64(&p.activeCount, 1)
				p.RunF(data)
				atomic.AddInt64(&p.activeCount, -1)
			} else if p.RunI != nil {
				atomic.AddInt64(&p.activeCount, 1)
				i := p.RunI
				i.Do(data)
				atomic.AddInt64(&p.activeCount, -1)
			}
		},
		jobChan: make(chan interface{}),
		quit:    make(chan bool),
	}
}

// Run 执行
func (t *Worker) Run() {
	// 先注册
	select {
	case t.pool.workerQueuePool <- t.jobChan:
		atomic.AddInt64(&t.pool.workerCount, 1)
		t.pool.workers.Store(fmt.Sprintf("%p", t), t)
	default:
		return
	}

	t.pool.wg.Add(1)
	go func() {
		defer func() {
			atomic.AddInt64(&t.pool.workerCount, -1)
			t.pool.workers.Delete(fmt.Sprintf("%p", t))
			t.pool.wg.Done()
		}()

		for {
			select {
			case data := <-t.jobChan:
				if data == nil {
					return
				}
				t.handler(data)
				tr := time.NewTimer(t.pool.IdleTimeout)
				select {
				case t.pool.workerQueuePool <- t.jobChan:
				case <-tr.C:
					return
				}
			case <-t.quit:
				close(t.jobChan)
			}
		}
	}()
}

// Stop 停止
func (t *Worker) Stop() {
	go func() {
		t.quit <- true
	}()
}
