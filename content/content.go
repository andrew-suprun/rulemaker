package content

import (
	"io/ioutil"
	"os"
	"strings"
)

type Content interface {
	Save() error
	Path() string
	Runes() [][]rune
	InsertLine(lineOffset int, line []rune)
	InsertLines(lineOffset int, lines [][]rune)
	RemoveLines(lineOffset, numLines int)
	InsertRune(lineOffset, runeOffset int, rune rune)
	InsertRunes(lineOffset, runeOffset int, runes []rune)
	RemoveRunes(lineOffset, runeOffset, numRunes int)
	SplitLine(lineOffset, runeOffset int)
	JoinLines(lineOffset int) int
}

func NewContent(path string) (Content, error) {
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
	return &content{
		path:  path,
		runes: runes,
	}, nil
}

type content struct {
	path  string
	runes [][]rune
}

func (c *content) Save() error {
	return nil // TODO
}

func (c *content) Path() string {
	return c.path
}

func (c *content) Runes() [][]rune {
	return c.runes
}

func (c *content) InsertLine(lineOffset int, line []rune) {
	// TODO
}

func (c *content) InsertLines(lineOffset int, lines [][]rune) {
	// TODO
}

func (c *content) RemoveLines(lineOffset, numLines int) {
	// TODO
}

func (c *content) InsertRune(lineOffset, runeOffset int, ch rune) {
	for len(c.runes) <= lineOffset {
		c.runes = append(c.runes, nil)
	}
	line := c.runes[lineOffset]
	for len(line) <= runeOffset {
		line = append(line, ' ')
	}
	rightPart := append([]rune{ch}, line[runeOffset:]...)
	line = append(line[:runeOffset], rightPart...)
	c.runes[lineOffset] = line
}

func (c *content) InsertRunes(lineOffset, runeOffset int, runes []rune) {
	// TODO
}

func (c *content) RemoveRunes(lineOffset, runeOffset, numRunes int) {
	if len(c.runes) <= lineOffset {
		return
	}
	line := c.runes[lineOffset]
	if len(line) <= runeOffset {
		return
	}
	if runeOffset >= len(line) {
		return
	}
	line = append(line[:runeOffset], line[runeOffset+numRunes:]...)
	c.runes[lineOffset] = line
}

func (c *content) SplitLine(lineOffset, runeOffset int) {
	if len(c.runes) <= lineOffset {
		return
	}
	line := c.runes[lineOffset]

	if runeOffset > len(line) {
		runeOffset = len(line)
	}
	line1 := append([]rune{}, line[:runeOffset]...)
	line2 := append([]rune{}, line[runeOffset:]...)
	result := make([][]rune, 0, 0)
	result = append(result, c.runes[:lineOffset]...)
	result = append(result, line1)
	result = append(result, line2)
	result = append(result, c.runes[lineOffset+1:]...)
	c.runes = result
}

func (c *content) JoinLines(lineOffset int) int {
	if lineOffset == 0 || len(c.runes) <= lineOffset {
		return 0
	}
	line := c.runes[lineOffset-1]
	line = append(line, c.runes[lineOffset]...)
	result := append(c.runes[:lineOffset-1], line)
	result = append(result, c.runes[lineOffset+1:]...)
	c.runes = result
	return len(line)
}
