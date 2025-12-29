package models

// для внешнего API сервиса
type AccrualResponse struct {
	Order   string
	Accrual int64
	Status  string
}
