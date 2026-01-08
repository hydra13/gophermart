package validators_test

import (
	"testing"

	"github.com/hydra13/gophermart/internal/validators"
	"github.com/stretchr/testify/assert"
)

func TestIsValidWithdrawAmount(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  bool
	}{
		{
			name:  "valid amount",
			input: 100,
			want:  true,
		},
		{
			name:  "invalid amount",
			input: -100,
			want:  false,
		},
		{
			name:  "invalid amount",
			input: 1_000_001_00,
			want:  false,
		},
		{
			name:  "zero amount",
			input: 0,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validators.IsValidWithdrawAmount(tt.input)

			assert.Equal(t, tt.want, got)
		})
	}
}
