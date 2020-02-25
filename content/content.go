package content

import (
	"io/ioutil"
	"os"
	"strings"

	"league.com/rulemaker/model"
)

type Content struct {
	Path      string
	Runes     [][]rune
	Cursor    model.Cursor
	Selection model.Selection
}

func NewContent(runes [][]rune) *Content {
	return &Content{Runes: runes}
}

func NewFileContent(path string) (*Content, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bytes), "\n")
	runes := make([][]rune, len(lines))
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", " ")
		line = strings.TrimRight(line, " ")
		runes[i] = []rune(line)
	}
	for len(runes[len(runes)-1]) == 0 {
		runes = runes[:len(runes)-1]
	}
	return &Content{
		Path:  path,
		Runes: runes,
	}, nil
}

func (c *Content) Save() error {
	return nil // TODO
}

func (c *Content) Columns(line int) int {
	return len(c.Runes[line])
}

func (c *Content) SetCursor(line, column int) {
	c.Cursor.Column = column
	if c.Cursor.Column < 0 {
		c.Cursor.Column = 0
	}
	c.Cursor.Line = line
	if c.Cursor.Line < 0 {
		c.Cursor.Line = 0
	}
}

func (c *Content) SetSelection(selection model.Selection) {
	c.Selection = selection
}

func (c *Content) MoveCursorUp(lines int) {
	c.Cursor.Line -= lines
	if c.Cursor.Line < 0 {
		c.Cursor.Line = 0
	}
}

func (c *Content) MoveCursorDown(lines, height int) {
	c.Cursor.Line += lines
	if c.Cursor.Line > len(c.Runes)+height-1 {
		c.Cursor.Line = len(c.Runes) + height - 1
	}
}

func (c *Content) MoveCursorLeft(columns int) {
	c.Cursor.Column -= columns
	if c.Cursor.Column < 0 {
		c.Cursor.Column = 0
	}
}

func (c *Content) MoveCursorRight(columns int) {
	c.Cursor.Column += columns
}

func (c *Content) MoveCursorToBol() {
	c.Cursor.Column = 0
}

func (c *Content) MoveCursorToEol() {
	if c.Cursor.Line < len(c.Runes) {
		c.Cursor.Column = len(c.Runes[c.Cursor.Line])
	} else {
		c.Cursor.Column = 0
	}

}

func (c *Content) InsertRune(ch rune) {
	c.InsertRunes([]rune{ch})
}

func (c *Content) InsertRunes(runes []rune) {
	for len(c.Runes) <= c.Cursor.Line {
		c.Runes = append(c.Runes, nil)
	}
	line := c.Runes[c.Cursor.Line]
	for len(line) <= c.Cursor.Column {
		line = append(line, ' ')
	}
	rightPart := append(runes, line[c.Cursor.Column:]...)
	line = append(line[:c.Cursor.Column], rightPart...)
	c.Runes[c.Cursor.Line] = line
	c.Cursor.Column += int(len(runes))
}

func (c *Content) DeleteLeft() {
	c.Cursor.Column--

	if c.Cursor.Column == -1 && c.Cursor.Line > 0 {
		line := c.Runes[c.Cursor.Line-1]
		column := len(line)
		line = append(line, c.Runes[c.Cursor.Line]...)
		runes := append(c.Runes[:c.Cursor.Line-1], line)
		runes = append(runes, c.Runes[c.Cursor.Line+1:]...)
		c.Runes = runes
		c.SetCursor(c.Cursor.Line-1, column)
		return
	}

	c.DeleteRight()
}

func (c *Content) DeleteRight() {
	if c.Cursor.Line >= len(c.Runes) {
		return
	}
	line := c.Runes[c.Cursor.Line]
	if c.Cursor.Column >= len(line) {
		return
	}
	line = append(line[:c.Cursor.Column], line[c.Cursor.Column+1:]...)
	c.Runes[c.Cursor.Line] = line
}

func (c *Content) SplitLine() {
	if len(c.Runes) <= c.Cursor.Line {
		return
	}
	line := c.Runes[c.Cursor.Line]

	if c.Cursor.Column > len(line) {
		c.Cursor.Column = len(line)
	}
	line1 := append([]rune{}, line[:c.Cursor.Column]...)
	line2 := append([]rune{}, line[c.Cursor.Column:]...)
	result := make([][]rune, 0, 0)
	result = append(result, c.Runes[:c.Cursor.Line]...)
	result = append(result, line1)
	result = append(result, line2)
	result = append(result, c.Runes[c.Cursor.Line+1:]...)
	c.Runes = result
	c.MoveCursorDown(1, 1)
	c.MoveCursorToBol()
}

func (c *Content) WrappedLines(width int) (result int) {
	for _, line := range c.Runes {
		result += len(splitLine(line, width))
	}
	return result
}

type RuneStream interface {
	Rune(ch rune, contentCursor, screenCursor model.Cursor)
	BreakRune(screenCursor model.Cursor)
	ContinueRune(screenCursor model.Cursor)
}

func (c *Content) StreamText(physicalLineStart, physicalLineEnd, width int, stream RuneStream) {
	physicalLineNum := 0
	screenLineNum := 0
	for lineIndex, line := range c.Runes {
		columnIndex := 0
		lines := splitLine(line, width)
		for i, wrappedLine := range lines {
			if physicalLineNum < physicalLineStart {
				physicalLineNum++
				continue
			}
			offset := 0
			if i > 0 {
				stream.ContinueRune(model.Cursor{Line: screenLineNum, Column: 3})
				offset = 4
			}

			for j, ch := range wrappedLine {
				stream.Rune(ch, model.Cursor{Line: lineIndex, Column: columnIndex}, model.Cursor{Line: screenLineNum, Column: offset + j})
				columnIndex++
			}

			if i < len(lines)-1 {
				stream.BreakRune(model.Cursor{Line: screenLineNum, Column: width - 1})
			}
			physicalLineNum++
			screenLineNum++
			if physicalLineNum >= physicalLineEnd {
				return
			}
		}
	}
}

func splitLine(line []rune, width int) [][]rune {
	if len(line) < width-1 {
		return [][]rune{line}
	}
	result := [][]rune{line[:width-1]}
	line = line[width-1:]
	for len(line) > 0 {
		if len(line) < width-5 {
			result = append(result, line)
			break
		}
		result = append(result, line[:width-5])
		line = line[width-5:]
	}
	return result
}
