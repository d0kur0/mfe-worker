package queue

import (
	"github.com/samber/lo"
	"log"
	"mfe-worker/internal/configMap"
	"sync"
	"time"
)

type Status uint

const (
	StatusLock Status = iota
	StatusFree        = iota
)

const Length = 5

type Worker func(wg *sync.WaitGroup) error

type Queue struct {
	queue       []Worker
	configMap   *configMap.ConfigMap
	queueStatus Status
}

func (q *Queue) StartQueueWorker() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				if q.queueStatus == StatusLock {
					continue
				}

				q.queueStatus = StatusLock

				var wg sync.WaitGroup
				var batch = lo.Slice(q.queue, 0, Length)

				for _, task := range batch {
					wg.Add(1)
					task := task
					go func() {
						err := task(&wg)
						if err != nil {
							log.Printf("queue task error: %s", err)
						}
					}()
				}

				wg.Wait()

				q.queue = lo.Slice(q.queue, Length, len(q.queue))
				q.queueStatus = StatusFree
			}
		}
	}()
}

func (q *Queue) AddToQueue(fn Worker) {
	q.queue = append(q.queue, fn)
}

func NewQueue(configMap *configMap.ConfigMap) *Queue {
	return &Queue{
		configMap:   configMap,
		queueStatus: StatusFree,
	}
}
