package view

import (
	"log"

	"github.com/gdamore/tcell"
	"league.com/rulemaker/model"
)

type View struct {
	Left, Width int
	Top, Height int
	LineOffset  int
	Style       tcell.Style
}

func NewView(style tcell.Style) *View {
	return &View{Style: style}
}

func (v *View) Resize(top, bottom, left, right int) {
	v.Top = top
	v.Height = bottom - top
	v.Left = left
	v.Width = right - left
}

func (v *View) ForCursor(y, x int, op func(line, column int)) {
	op(y+v.LineOffset-v.Top, x-v.Left)
}

func (v *View) ForSelection(selection model.Selection, op func(y, x int)) {
	if selection.End.Line < v.LineOffset || selection.Start.Line > v.LineOffset+v.Height {
		return
	}
	for line := selection.Start.Line; line <= selection.End.Line; line++ {
		if line < v.LineOffset || line > v.LineOffset+v.Height {
			continue
		}

		startColumn, endColumn := 0, v.Width
		if line == selection.Start.Line {
			startColumn = selection.Start.Column
		}
		if line == selection.End.Line {
			endColumn = selection.End.Column
		}
		log.Printf("### line=%d  columns=%d:%d\n", line, startColumn, endColumn)

		y := line + v.Top - v.LineOffset
		for column := startColumn; column < endColumn; column++ {
			x := column + v.Left
			op(y, x)
		}
	}
}

func (v *View) MakeCursorVisible(line, column int) {
	if v.LineOffset > line {
		v.LineOffset = line
	}
	if v.LineOffset <= line-v.Height {
		v.LineOffset = line - v.Height + 1
	}
}

func (v *View) Contains(y, x int) bool {
	return x >= v.Left && x < v.Left+v.Width && y >= v.Top && y < v.Top+v.Height
}

func (v *View) Scroll(lines, maxLineOffset int) {
	v.LineOffset += lines
	if v.LineOffset > maxLineOffset {
		v.LineOffset = maxLineOffset
	}
	if v.LineOffset < 0 {
		v.LineOffset = 0
	}
}

func (v *View) WrappedLines(runes [][]rune) (result int) {
	for _, line := range runes {
		result += len(splitLine(line, v.Width))
	}
	return result
}

type LineNumberStream interface {
	Line(contentLine, screenLine int)
}

func (v *View) StreamLines(runes [][]rune, width int, stream LineNumberStream) {
	physicalLineNum := 0
	for lineIndex, line := range runes {
		lines := splitLine(line, width)
		if physicalLineNum < v.LineOffset {
			physicalLineNum += len(lines)
			continue
		}
		if physicalLineNum-v.LineOffset+v.Top > v.Height+1 {
			return
		}
		stream.Line(lineIndex, physicalLineNum-v.LineOffset+v.Top)
		physicalLineNum += len(lines)
	}
}

type RuneStream interface {
	Rune(ch rune, contentCursor, screenCursor model.Cursor)
	LineBreak(screenCursor model.Cursor)
}

func (v *View) StreamText(runes [][]rune, stream RuneStream) {
	physicalLineNum := 0
	for lineIndex, line := range runes {
		columnIndex := 0
		lines := splitLine(line, v.Width)
		for i, wrappedLine := range lines {
			if physicalLineNum < v.LineOffset {
				physicalLineNum++
				continue
			}

			for j, ch := range wrappedLine {
				stream.Rune(ch, model.Cursor{Line: lineIndex, Column: columnIndex}, model.Cursor{Line: physicalLineNum - v.LineOffset + v.Top, Column: j + v.Left})
				columnIndex++
			}

			if i < len(lines)-1 {
				stream.LineBreak(model.Cursor{Line: physicalLineNum - v.LineOffset + v.Top, Column: v.Width - 1 + v.Left})
			}
			physicalLineNum++
			if physicalLineNum >= v.LineOffset+v.Height {
				return
			}
		}
	}
}

func splitLine(line []rune, width int) [][]rune {
	result := make([][]rune, 0, len(line)/(width-1)+1)
	for len(line) >= width {
		result = append(result, line[:width-1])
		line = line[width-1:]
	}
	result = append(result, line)
	return result
}
