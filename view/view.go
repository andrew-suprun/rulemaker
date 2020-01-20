package view

import (
	"github.com/gdamore/tcell"
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

func (v *View) ClipText(txt string, line, column int) (string, int, int) {
	if line < v.LineOffset || line >= v.LineOffset+v.Height ||
		column+len(txt) < v.ColumnOffset || column >= v.ColumnOffset+v.Width {

		return "", -1, -1
	}
	if column+len(txt) > v.ColumnOffset+v.Width {
		txt = txt[:v.ColumnOffset+v.Width-column]
	}
	if column < v.ColumnOffset {
		txt = txt[v.ColumnOffset-column:]
		column = v.ColumnOffset
	}
	line += v.Top - v.LineOffset
	column += v.Left - v.ColumnOffset
	return txt, line, column
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

func (v *View) Contains(physicalLine, physicalColumn int) bool {
	return physicalColumn >= v.Left && physicalColumn < v.Left+v.Width && physicalLine >= v.Top && physicalLine < v.Top+v.Height
}

func (v *View) CursorFromScreenCoordinates(physicalLine, physicalColumn int) (cursorLine, cursorColumn int) {
	return physicalLine + v.LineOffset - v.Top, physicalColumn + v.ColumnOffset - v.Left
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
