package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  bool
	}{
		{
			name:  "success",
			input: 4440,
			want:  true,
		},
		{
			name:  "not valid",
			input: 4444,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidLuhn(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
