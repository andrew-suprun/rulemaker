package view

import (
	"log"

	"github.com/gdamore/tcell"
	"league.com/rulemaker/model"
)

type View struct {
	Left, Width              int
	Top, Height              int
	ColumnOffset, LineOffset int
	Style                    tcell.Style
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
	op(y+v.LineOffset-v.Top, x+v.ColumnOffset-v.Left)
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
			x := column + v.Left - v.ColumnOffset
			op(y, x)
		}
	}
}

func (v *View) ShowText(txt string, line, column int, op func(y, x int, ch rune)) {
	if line < v.LineOffset || line >= v.LineOffset+v.Height ||
		column+len(txt) < v.ColumnOffset || column >= v.ColumnOffset+v.Width {
		return
	}
	if column+len(txt) > v.ColumnOffset+v.Width {
		txt = txt[:v.ColumnOffset+v.Width-column]
	}
	if column < v.ColumnOffset {
		txt = txt[v.ColumnOffset-column:]
		column = v.ColumnOffset
	}
	y := line + v.Top - v.LineOffset
	x := column + v.Left - v.ColumnOffset
	for i, ch := range txt {
		op(y, x+i, ch)
	}
}

func (v *View) MakeCursorVisible(line, column int) {
	if v.LineOffset > line {
		v.LineOffset = line
	}
	if v.LineOffset <= line-v.Height {
		v.LineOffset = line - v.Height + 1
	}
	if v.ColumnOffset > column {
		v.ColumnOffset = column
	}
	if v.ColumnOffset <= column-v.Width {
		v.ColumnOffset = column - v.Width + 1
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
