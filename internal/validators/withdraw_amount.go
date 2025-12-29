package validators

func IsValidWithdrawAmount(amount int64) bool {
	return amount > 0 && amount < 1_000_000_00
}
