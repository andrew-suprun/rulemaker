package viewport

import (
	"testing"
)

var fixtures = []struct {
	text     string
	width    int
	expected int
}{
	{"", 8, 1},
	{"1234567", 8, 1},
	{"12345678", 8, 2},
	{"1234567890", 8, 2},
	{"12345678901", 8, 3},
}

func TestScreenLines(t *testing.T) {
	for _, fixture := range fixtures {
		result := ScreenLines([][]rune{[]rune(fixture.text)}, fixture.width)
		if result != fixture.expected {
			t.Fatalf("text %q, expected %d, got %d", fixture.text, fixture.expected, result)
		}
	}
	var lines [][]rune
	expected := 0
	for _, fixture := range fixtures {
		lines = append(lines, []rune(fixture.text))
		expected += fixture.expected
	}
	result := ScreenLines(lines, 8)
	if result != expected {
		t.Fatalf("expected %d, got %d", expected, result)
	}
}
