package viewport

func ScreenLines(text [][]rune, width int) (result int) {
	for _, line := range text {
		if len(line) < width-1 {
			result++
		} else {
			result += (len(line)-5)/(width-5) + 1
		}
	}
	return result
}
