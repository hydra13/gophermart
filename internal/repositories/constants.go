package repositories

import "time"

const (
	// RetryMaxRetries максимальное количество попыток повтора
	RetryMaxRetries = 3

	// RetryBaseDelay базовая задержка между попытками (экспоненциальная задержка)
	RetryBaseDelay = time.Millisecond * 100

	// RetryMaxDelay максимальная задержка между попытками
	RetryMaxDelay = time.Second
)
