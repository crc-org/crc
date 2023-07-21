package strings

func Contains(input []string, match string) bool {
	for _, v := range input {
		if v == match {
			return true
		}
	}
	return false
}
