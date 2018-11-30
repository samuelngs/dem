package ext

import "time"

// ProgressBar interface
type ProgressBar interface {
	IncrBy(int, ...time.Duration)
	Completed() bool
}
