package validators

import "strings"

func IsValidPass(password string) bool {
	if len(password) < 8 {
		return false
	}
	if len(password) > 128 {
		return false
	}
	if strings.Contains(password, " ") {
		return false
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return false
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return false
	}
	if !strings.ContainsAny(password, "0123456789") {
		return false
	}
	if !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		return false
	}

	return true
}
