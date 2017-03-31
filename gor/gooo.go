package gor

import (
	"github.com/ivpusic/grpool"
)

func GoGoGo(workers int, queueSize int, jobs int, fn func(int)) {
	pool := grpool.NewPool(workers, queueSize)
	defer pool.Release()
	for i := 0; i < jobs; i++ {
		pool.JobQueue <- func(index int) func() {
			pool.WaitCount(1)
			return func() {
				defer pool.JobDone()
				fn(index)
			}
		}(i)
	}
	pool.WaitAll()
}
