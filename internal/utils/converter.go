package utils

func ToRub(sum int64) float64 {
	return float64(sum) / 100
}

func FromRub(sum float64) int64 {
	return int64(sum * 100)
}
