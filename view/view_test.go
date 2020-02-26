package view

import (
	"fmt"
	"reflect"
	"testing"

	"league.com/rulemaker/model"
)

var fixtures = []struct {
	text     string
	expected int
}{
	{"", 1},
	{"1234567", 1},
	{"12345678", 2},
	{"1234567890", 2},
	{"12345678901", 3},
}

func TestWrappedLines(t *testing.T) {
	v := NewView(0)
	v.Resize(0, 10, 0, 10)
	v.Width = 8

	for _, fixture := range fixtures {
		result := v.WrappedLines([][]rune{[]rune(fixture.text)})
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
	result := v.WrappedLines(lines)
	if result != expected {
		t.Fatalf("expected %d, got %d", expected, result)
	}
}

func TestStreamRunes(t *testing.T) {
	testStreamRunesSubrange(t, 0, 9)
	testStreamRunesSubrange(t, 1, 8)
	testStreamRunesSubrange(t, 2, 7)
	testStreamRunesSubrange(t, 3, 6)
	testStreamRunesSubrange(t, 4, 5)
	testStreamRunesSubrange(t, 5, 6)
}

func testStreamRunesSubrange(t *testing.T, start, end int) {
	v := NewView(0)
	v.Resize(0, end-start, 0, 8)
	v.LineOffset = start

	var lines [][]rune
	for _, fixture := range fixtures {
		lines = append(lines, []rune(fixture.text))
	}

	stream := newTestStream()
	v.StreamText(lines, stream)
	if !reflect.DeepEqual(stream.result[:end-start], runesExpected[start:end]) {
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

var runesExpected = [][]rune{
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

func TestStreamLineNumbers(t *testing.T) {
	v := NewView(0)
	v.Resize(0, 10, 0, 8)

	var lines [][]rune
	for _, fixture := range fixtures {
		lines = append(lines, []rune(fixture.text))
	}

	stream := &testLineNumbersStream{}
	v.StreamLines(lines, stream)
	if !reflect.DeepEqual(stream.result, linesExpected) {
		t.FailNow()
	}
}

type testLineNumbersStream struct {
	result []pair
}

type pair struct {
	contentLine int
	screenLine  int
}

var linesExpected = []pair{
	{0, 0},
	{1, 1},
	{2, 2},
	{3, 4},
	{4, 6},
}

func (s *testLineNumbersStream) Line(contentLine, screenLine int) {
	s.result = append(s.result, pair{contentLine: contentLine, screenLine: screenLine})
}
