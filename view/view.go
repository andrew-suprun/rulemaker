package view

import (
	"github.com/gdamore/tcell"
)

type View interface {
	Width() int
	Height() int
	SetOffsets(x, y int)
	Clear(style tcell.Style)
	SetText(text string, x, y int, style tcell.Style)
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
