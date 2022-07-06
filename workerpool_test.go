package xwp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

var (
	count int64
)

type TestWorker struct {
}

func (t *TestWorker) Do(data interface{}) {
	atomic.AddInt64(&count, 1)
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
}

func TestOnceRun(t *testing.T) {
	a := assert.New(t)

	jobQueue := make(chan interface{}, 200)
	num := 10000000
	p := &WorkerPool{
		JobQueue:       jobQueue,
		MaxWorkers:     1000,
		InitWorkers:    100,
		MaxIdleWorkers: 100,
		RunI:           &TestWorker{},
	}

	go func() {
		for i := 0; i < num; i++ {
			jobQueue <- i
		}
		p.Stop()
	}()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			<-ticker.C
			log.Printf("%+v", p.Stats())
		}
	}()

	p.Run()

	a.Equal(count, int64(num))
}

func TestStop(t *testing.T) {
	a := assert.New(t)

	jobQueue := make(chan interface{}, 200)
	p := &WorkerPool{
		JobQueue:   jobQueue,
		MaxWorkers: 10,
		RunI:       &TestWorker{},
	}

	go func() {
		defer func() {
			err := recover()
			a.EqualError(err.(error), "send on closed channel")
		}()
		for {
			jobQueue <- struct {
			}{}
		}
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		p.Stop()
	}()

	p.Run()
}

type TestWorkerRunF struct {
	i int
}

func (tw *TestWorkerRunF) Do() {
	time.Sleep(time.Second)
	fmt.Println("序号", tw.i, "完成")
}

func TestRunF(t *testing.T) {
	jobQueue := make(chan interface{}, 200)
	p := &WorkerPool{
		JobQueue:       jobQueue,
		MaxWorkers:     100,
		InitWorkers:    10,
		MaxIdleWorkers: 10,
		IdleTimeout:    10 * time.Second,
	}

	go func() {
		time.Sleep(time.Second)
		// 投放任务
		for i := 0; i < 20; i++ {
			tw := &TestWorkerRunF{i: i}
			jobQueue <- tw
		}

		// 投放完停止调度
		//p.Stop()
	}()

	p.Start()

	for {
		t.Log(time.Now(), p.Stats())
		time.Sleep(500 * time.Millisecond)
	}
}

func TestRunFunc(t *testing.T) {
	jobQueue := make(chan interface{}, 200)
	p := &WorkerPool{
		JobQueue:       jobQueue,
		MaxWorkers:     100,
		InitWorkers:    10,
		MaxIdleWorkers: 10,
		IdleTimeout:    10 * time.Second,
	}

	go func() {
		// 投放任务
		for i := 0; i < 20; i++ {
			tw := &TestWorkerRunF{i: i}
			jobQueue <- RunFunc(func() {
				time.Sleep(time.Second)
				t.Log("序号", tw.i, "完成")
			})
		}

		// 投放完停止调度
		//p.Stop()
	}()

	p.Start()

	for {
		t.Log(time.Now(), p.Stats())
		time.Sleep(500 * time.Millisecond)
	}
}
