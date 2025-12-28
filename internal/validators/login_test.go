package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidLogin(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid login",
			input: "login",
			want:  true,
		},
		{
			name:  "invalid login - empty",
			input: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidLogin(tt.input)
			assert.Equal(t, got, tt.want)
		})
	}
}
