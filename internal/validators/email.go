package validators

import (
	"net/mail"
	"regexp"
)

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// доп проверка для тех случаев, когда net/mail считает что все валидно (например, "test@localhost")
	re, err := regexp.Compile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		// возвращаем true, так как основная валидация через net/mail уже прошла успешно
		return true
	}

	return re.MatchString(email)
}
