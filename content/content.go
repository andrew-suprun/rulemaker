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
	InsertLines(lineOffset int, lines [][]rune)
	RemoveLines(lineOffset, numLines int)
	InsertRunes(lineOffset, runeOffset int, runes []rune)
	RemoveRunes(lineOffset, runeOffset, numRunes int)
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

func (c *content) InsertLines(lineOffset int, lines [][]rune) {
	// TODO
}

func (c *content) RemoveLines(lineOffset, numLines int) {
	// TODO
}

func (c *content) InsertRunes(lineOffset, runeOffset int, runes []rune) {
	// TODO
}

func (c *content) RemoveRunes(lineOffset, runeOffset, numRunes int) {
	// TODO
}
