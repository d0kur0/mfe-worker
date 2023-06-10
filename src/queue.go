package src

import (
	"github.com/samber/lo"
	"sync"
	"time"
)

type QueueStatus uint

const (
	QueueStatusLock QueueStatus = iota
	QueueStatusFree             = iota
)

const QueueLength = 5

type Worker func(wg *sync.WaitGroup)

type Queue struct {
	queue       []Worker
	configMap   *ConfigMap
	queueStatus QueueStatus
}

func (q *Queue) StartQueueWorker() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				if q.queueStatus == QueueStatusLock {
					continue
				}

				q.queueStatus = QueueStatusLock

				var wg sync.WaitGroup
				var batch = lo.Slice(q.queue, 0, QueueLength)

				for _, task := range batch {
					wg.Add(1)
					go task(&wg)
				}

				wg.Wait()

				q.queue = lo.Slice(q.queue, QueueLength, len(q.queue))
				q.queueStatus = QueueStatusFree
			}
		}
	}()
}

func (q *Queue) AddToQueue(fn Worker) {
	q.queue = append(q.queue, fn)
}

func NewQueue(configMap *ConfigMap) *Queue {
	return &Queue{
		configMap:   configMap,
		queueStatus: QueueStatusFree,
	}
}
