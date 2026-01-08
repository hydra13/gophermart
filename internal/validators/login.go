package validators

import "strings"

func IsValidLogin(login string) bool {
	if len(login) < 3 || len(login) > 50 {
		return false
	}

	if strings.Contains(login, " ") {
		return false
	}

	return true
}
