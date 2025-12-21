package utils

import (
	"testing"
)

func TestToRub(t *testing.T) {
	tests := []struct {
		name string
		sum  int64
		want float64
	}{
		{
			name: "Целые",
			sum:  10000,
			want: 100.00,
		},
		{
			name: "С дробной частью",
			sum:  10001,
			want: 100.01,
		},
		{
			name: "Ноль",
			sum:  0,
			want: 0.00,
		},
		{
			name: "Отрицательное значение",
			sum:  -5000,
			want: -50.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToRub(tt.sum); got != tt.want {
				t.Errorf("ToRub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromRub(t *testing.T) {
	tests := []struct {
		name string
		sum  float64
		want int64
	}{
		{
			name: "Целые",
			sum:  100.00,
			want: 10000,
		},
		{
			name: "С дробной частью",
			sum:  100.01,
			want: 10001,
		},
		{
			name: "С большой дробной частью",
			sum:  100.017,
			want: 10001,
		},
		{
			name: "Ноль",
			sum:  0.00,
			want: 0,
		},
		{
			name: "Отрицательное значение",
			sum:  -50.00,
			want: -5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromRub(tt.sum); got != tt.want {
				t.Errorf("FromRub() = %v, want %v", got, tt.want)
			}
		})
	}
}
