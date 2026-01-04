package repositories

import "time"

const (
	RetryMaxRetries = 3
	RetryBaseDelay  = time.Millisecond * 100
	RetryMaxDelay   = time.Second
)
