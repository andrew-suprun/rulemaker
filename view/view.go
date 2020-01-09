package view

import (
	"github.com/gdamore/tcell"
)

type View interface {
	Width() int
	Height() int
	SetOffsets(x, y int)
	Offsets() (x, y int)
	Clear(style tcell.Style)
	SetText(text string, x, y int, style tcell.Style)
	ShowCursor(x, y int)
	Contains(physicalX, physicalY int) bool
	CursorFromPhysicalCoordinates(physicalX, physicalY int) (cursorX, cursorY int)
}

func NewView(screen tcell.Screen, left, right, top, bottom int) View {
	return &view{
		screen: screen,
		left:   left,
		width:  right - left,
		top:    top,
		height: bottom - top,
	}
}

type view struct {
	screen                   tcell.Screen
	left, width, top, height int
	style                    tcell.Style
	offsetX, offsetY         int
}

func (v *view) Width() int {
	return v.width
}

func (v *view) Height() int {
	return v.height
}

func (v *view) Clear(style tcell.Style) {
	for row := 0; row < v.height; row++ {
		for col := 0; col < v.width; col++ {
			v.screen.SetContent(col+v.left, row+v.top, ' ', nil, style)
		}
	}
}

func (v *view) SetOffsets(x, y int) {
	v.offsetX = x
	v.offsetY = y
}

func (v *view) Offsets() (x, y int) {
	return v.offsetX, v.offsetY
}

func (v *view) SetText(text string, x, y int, style tcell.Style) {
	x = x - v.offsetX
	y = y - v.offsetY
	if x+len(text) < 0 {
		return
	}
	if x >= +v.width {
		return
	}
	if y < 0 {
		return
	}
	if y >= v.height {
		return
	}
	leftCut, rightCut := 0, len(text)
	if x < 0 {
		leftCut = -x
		x = 0
	}
	if x+len(text) >= v.width {
		rightCut = v.width - x
	}

	text = text[leftCut:rightCut]

	x = x + v.left
	y = y + v.top
	for _, c := range text {
		v.screen.SetContent(x, y, c, nil, style)
		x++
	}
}

func (v *view) ShowCursor(x, y int) {
	if x < v.offsetX || y < v.offsetY || x >= v.offsetX+v.width || y >= v.offsetY+v.height {
		v.screen.HideCursor()
		return
	}
	v.screen.ShowCursor(x-v.offsetX+v.left, y-v.offsetY+v.top)
}

func (v *view) Contains(physicalX, physicalY int) bool {
	return physicalX >= v.left && physicalX < v.left+v.width && physicalY >= v.top && physicalY < v.top+v.height
}

func (v *view) CursorFromPhysicalCoordinates(physicalX, physicalY int) (cursorX, cursorY int) {
	return physicalX + v.offsetX - v.left, physicalY + v.offsetY - v.top
}
