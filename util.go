package rush

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func indexToLabelRune(i int) rune {
	return 'A' + rune(i)
}

func indexToLabelString(i int) string {
	return string(indexToLabelRune(i))
}
