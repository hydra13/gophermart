package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid email",
			input: "test@example.com",
			want:  true,
		},
		{
			name:  "invalid email 1",
			input: "test@localhost",
			want:  false,
		},
		{
			name:  "invalid email 2",
			input: "testexample.com",
			want:  false,
		},
		{
			name:  "invalid email 3",
			input: "@example.com",
			want:  false,
		},
		{
			name:  "empty email",
			input: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
