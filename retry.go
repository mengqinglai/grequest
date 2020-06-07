package grequest

import "time"

type RetryConfig struct {
	Status      map[int]struct{}
	StatusBegin int
	StatusEnd   int
	Interval    time.Duration
	MaxRetry    int
	Counter     int
}

func NewRetryConfig(maxRetry int, interval time.Duration, status []int, statusBegin int, statusEnd int) *RetryConfig {
	rc := &RetryConfig{MaxRetry: maxRetry, Interval: interval, StatusBegin: statusBegin, StatusEnd: statusEnd}
	rc.Status = make(map[int]struct{}, len(status))
	for _, st := range status {
		rc.Status[st] = struct{}{}
	}
	return rc
}

func (rc *RetryConfig) isRetryStatus(status int) bool {
	if rc.StatusBegin <= status && status <= rc.StatusEnd {
		return true
	}
	_, ok := rc.Status[status]
	return ok
}

func (rc *RetryConfig) isRetry(err error, status int) bool {
	if rc == nil {
		return false
	}
	if (err != nil || rc.isRetryStatus(status)) && rc.Counter < rc.MaxRetry {
		return true
	}
	return false
}

func (rc *RetryConfig) isContinueRetry(err error, status int) bool {
	if rc.isRetry(err, status) {
		rc.Counter++
		time.Sleep(rc.Interval)
		return true
	}
	return false
}
