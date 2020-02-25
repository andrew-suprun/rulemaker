package content

import (
	"fmt"
	"reflect"
	"testing"

	"league.com/rulemaker/model"
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

func TestWrappedLines(t *testing.T) {
	for _, fixture := range fixtures {
		content := NewContent([][]rune{[]rune(fixture.text)})
		result := content.WrappedLines(fixture.width)
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
	context := NewContent(lines)
	result := context.WrappedLines(8)
	if result != expected {
		t.Fatalf("expected %d, got %d", expected, result)
	}
}

func TestStreamRunes(t *testing.T) {
	var lines [][]rune
	for _, fixture := range fixtures {
		lines = append(lines, []rune(fixture.text))
	}
	content := NewContent(lines)
	testStreamRunesSubrange(t, 0, 9, content)
	testStreamRunesSubrange(t, 1, 8, content)
	testStreamRunesSubrange(t, 2, 7, content)
	testStreamRunesSubrange(t, 3, 6, content)
	testStreamRunesSubrange(t, 4, 5, content)
	testStreamRunesSubrange(t, 5, 6, content)
}

func testStreamRunesSubrange(t *testing.T, start, end int, content *Content) {
	stream := newTestStream()
	content.StreamText(start, end, 8, stream)
	if !reflect.DeepEqual(stream.result[:end-start], expected[start:end]) {
		fmt.Println(start, end, stream.result[:end-start])
		t.FailNow()
	}
}

type testStream struct {
	result [][]rune
}

func newTestStream() *testStream {
	result := make([][]rune, 9)
	for i := 0; i < 9; i++ {
		result[i] = []rune{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
	}
	return &testStream{result: result}
}

func (s *testStream) Rune(ch rune, contentCursor, screenCursor model.Cursor) {
	s.result[screenCursor.Line][screenCursor.Column] = ch
}

func (s *testStream) BreakRune(screenCursor model.Cursor) {
	s.result[screenCursor.Line][screenCursor.Column] = '↓'
}

func (s *testStream) ContinueRune(screenCursor model.Cursor) {
	s.result[screenCursor.Line][screenCursor.Column] = '→'
}

var expected = [][]rune{
	{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	{'1', '2', '3', '4', '5', '6', '7', ' '},
	{'1', '2', '3', '4', '5', '6', '7', '↓'},
	{' ', ' ', ' ', '→', '8', ' ', ' ', ' '},
	{'1', '2', '3', '4', '5', '6', '7', '↓'},
	{' ', ' ', ' ', '→', '8', '9', '0', ' '},
	{'1', '2', '3', '4', '5', '6', '7', '↓'},
	{' ', ' ', ' ', '→', '8', '9', '0', '↓'},
	{' ', ' ', ' ', '→', '1', ' ', ' ', ' '},
}
