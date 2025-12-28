package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPassValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid password",
			input: "ABcd1234",
			want:  true,
		},
		{
			name:  "invalid password",
			input: "ABcd123",
			want:  false,
		},
		{
			name:  "invalid password",
			input: "abc",
			want:  false,
		},
		{
			name:  "invalid password",
			input: "Testing.",
			want:  false,
		},
		{
			name:  "invalid password",
			input: "Testing ",
			want:  false,
		},
		{
			name:  "empty password",
			input: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPass(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
