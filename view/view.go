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

func (v *View) Resize(left, right, top, bottom int) {
	v.Left = left
	v.Width = right - left
	v.Top = top
	v.Height = bottom - top
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

func (v *View) SetCursor(column, line int) {
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
	v.SetCursor(column+v.CursorColumn, line+v.CursorLine)
}

func (v *View) Contains(physicalX, physicalY int) bool {
	return physicalX >= v.Left && physicalX < v.Left+v.Width && physicalY >= v.Top && physicalY < v.Top+v.Height
}

func (v *View) CursorFromScreenCoordinates(physicalX, physicalY int) (cursorX, cursorY int) {
	return physicalX + v.ColumnOffset - v.Left, physicalY + v.LineOffset - v.Top
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
