package view

import (
	"github.com/gdamore/tcell"
)

type View struct {
	Left, Width              int
	Top, Height              int
	ColumnOffset, LineOffset int
	CursorColumn, CursorLine int
	TotalLines               int
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
	if v.TotalLines <= line {
		v.TotalLines = line + 1
	}
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

func (v *View) SetCursor(line, column int) {
	v.CursorColumn = column
	if v.CursorColumn < 0 {
		v.CursorColumn = 0
	}
	if v.ColumnOffset > v.CursorColumn {
		v.ColumnOffset = v.CursorColumn
	}
	if v.ColumnOffset <= v.CursorColumn-v.Width {
		v.ColumnOffset = v.CursorColumn - v.Width + 1
	}
	v.CursorLine = line
	if v.CursorLine < 0 {
		v.CursorLine = 0
	}
	if v.CursorLine >= v.TotalLines+v.Height-1 {
		v.CursorLine = v.TotalLines + v.Height - 2
	}
	if v.LineOffset > v.CursorLine {
		v.LineOffset = v.CursorLine
	}
	if v.LineOffset <= v.CursorLine-v.Height {
		v.LineOffset = v.CursorLine - v.Height + 1
	}
}

func (v *View) MoveCursor(column, line int) {
	v.SetCursor(line+v.CursorLine, column+v.CursorColumn)
}

func (v *View) Contains(physicalLine, physicalColumn int) bool {
	return physicalColumn >= v.Left && physicalColumn < v.Left+v.Width && physicalLine >= v.Top && physicalLine < v.Top+v.Height
}

func (v *View) CursorFromScreenCoordinates(physicalLine, physicalColumn int) (cursorLine, cursorColumn int) {
	return physicalLine + v.LineOffset - v.Top, physicalColumn + v.ColumnOffset - v.Left
}

func (v *View) Scroll(lines int) {
	v.LineOffset += lines
	if v.LineOffset < 0 {
		v.LineOffset = 0
	} else if v.LineOffset >= v.TotalLines {
		v.LineOffset = v.TotalLines - 1
	}
}

func (v *View) PageUp() {
	v.Scroll(-v.Height)
	v.MoveCursor(0, -v.Height)
}

func (v *View) PageDown() {
	v.Scroll(v.Height)
	v.MoveCursor(0, v.Height)
}
