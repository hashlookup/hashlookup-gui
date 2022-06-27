package main

import (
	"github.com/Jeffail/tunny"
	"runtime"
)

var (
	_defaultWorkersNumber = runtime.NumCPU()
)

// TunnyJob holds a pool of workers that run a function in parallel
type TunnyJob struct {
	// tunny pool of worker
	Pool *tunny.Pool
	// Number of workers
	workersNumber int
	// the function to run on each workers
	jobFunc func()
	// the total number of job to be completed
	totalJobNumber int
}

func newTunnyJob(totalJobNumber int, jobFunc func(interface{}) interface{}, opt_numWorkers ...int) *TunnyJob {
	var workerNumber int

	if len(opt_numWorkers) > 0 {
		workerNumber = opt_numWorkers[0]
	} else {
		workerNumber = _defaultWorkersNumber
	}

	pool := tunny.NewFunc(workerNumber, jobFunc)

	return &TunnyJob{Pool: pool, workersNumber: workerNumber, totalJobNumber: totalJobNumber}
}
