package view

import (
	"github.com/gdamore/tcell"
)

type View interface {
	Resize(left, right, top, bottom int)
	LineOffset() int
	TotalLines() int
	Width() int
	Height() int
	Clear()
	Cursor() (column, line int)
	SetCursor(column, line int)
	Offsets() (column, line int)
	SetOffsets(column, line int)
	ShowCursor()
	SetText(text string, x, y int, style tcell.Style)
	Contains(physicalX, physicalY int) bool
	CursorFromScreenCoordinates(physicalX, physicalY int) (cursorX, cursorY int)
	Scroll(lines int)
	PageUp()
	PageDown()
}

type view struct {
	screen                   tcell.Screen
	style                    tcell.Style
	left, width, top, height int
	columnOffset, lineOffset int
	cursorColumn, cursorLine int
	totalLines               int
}

func NewView(screen tcell.Screen, style tcell.Style) View {
	return &view{screen: screen, style: style}
}

func (v *view) Resize(left, right, top, bottom int) {
	v.left = left
	v.width = right - left
	v.top = top
	v.height = bottom - top
}

func (v *view) Clear() {
	for row := 0; row < v.height; row++ {
		for col := 0; col < v.width; col++ {
			v.screen.SetContent(col+v.left, row+v.top, ' ', nil, v.style)
		}
	}
}

func (v *view) LineOffset() int {
	return v.lineOffset
}

func (v *view) TotalLines() int {
	return v.totalLines
}

func (v *view) Width() int {
	return v.width
}

func (v *view) Height() int {
	return v.height
}

func (v *view) Cursor() (column, line int) {
	return v.cursorColumn, v.cursorLine
}

func (v *view) SetCursor(column, line int) {
	v.cursorColumn = column
	if v.cursorColumn < 0 {
		v.cursorColumn = 0
	}
	if v.columnOffset > v.cursorColumn {
		v.columnOffset = v.cursorColumn
	}
	if v.columnOffset <= v.cursorColumn-v.width {
		v.columnOffset = v.cursorColumn - v.width + 1
	}
	v.cursorLine = line
	if v.cursorLine < 0 {
		v.cursorLine = 0
	}
	if v.lineOffset > v.cursorLine {
		v.lineOffset = v.cursorLine
	}
	if v.lineOffset <= v.cursorLine-v.height {
		v.lineOffset = v.cursorLine - v.height + 1
	}
}

func (v *view) Offsets() (column, line int) {
	return v.columnOffset, v.lineOffset
}

func (v *view) SetOffsets(column, line int) {
	v.columnOffset = column
	v.lineOffset = line
}

func (v *view) ShowCursor() {
	if v.cursorColumn < v.columnOffset || v.cursorLine < v.lineOffset || v.cursorColumn >= v.columnOffset+v.width || v.cursorLine >= v.lineOffset+v.height {
		v.screen.HideCursor()
		return
	}
	v.screen.ShowCursor(v.cursorColumn-v.columnOffset+v.left, v.cursorLine-v.lineOffset+v.top)
}

func (v *view) SetText(text string, x, y int, style tcell.Style) {
	if v.totalLines <= y {
		v.totalLines = y + 1
	}
	x = x - v.columnOffset
	y = y - v.lineOffset
	if x+len(text) < 0 {
		return
	}
	if x >= v.width {
		return
	}
	if y < 0 {
		return
	}
	if y >= v.height {
		return
	}
	if x < 0 {
		text = text[-x:]
		x = 0
	}

	if len(text) >= v.width-x {
		text = text[:v.width-x]
	}

	x = x + v.left
	y = y + v.top
	for _, c := range text {
		v.screen.SetContent(x, y, c, nil, style)
		x++
	}
}

func (v *view) Contains(physicalX, physicalY int) bool {
	return physicalX >= v.left && physicalX < v.left+v.width && physicalY >= v.top && physicalY < v.top+v.height
}

func (v *view) CursorFromScreenCoordinates(physicalX, physicalY int) (cursorX, cursorY int) {
	return physicalX + v.columnOffset - v.left, physicalY + v.lineOffset - v.top
}

func (v *view) Scroll(lines int) {
	v.lineOffset += lines
	if v.lineOffset < 0 {
		v.lineOffset = 0
	} else if v.lineOffset >= v.totalLines {
		v.lineOffset = v.totalLines - 1
	}
}

func (v *view) PageUp() {
	v.Scroll(-v.height)
}

func (v *view) PageDown() {
	v.Scroll(v.height)
}
