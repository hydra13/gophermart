package validators

import (
	"strconv"
)

func IsValidLuhn(number string) bool {
	nDigits := len(number)
	sum := 0
	isSecond := false

	for i := nDigits - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if isSecond {
			digit = digit * 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
