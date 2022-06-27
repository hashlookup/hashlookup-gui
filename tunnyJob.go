package main

import "github.com/Jeffail/tunny"

var (
	_defaultWorkersNumber = 4
)

// TunnyJob holds a pool of workers that run a function in parallel
type TunnyJob struct {
	// tunny pool of worker
	pool tunny.Pool
	// Number of workers
	workersNumber int
	// the function to run on each workers
	jobFunc func()
	// the total number of job to be completed
	totalJobNumber int
}
