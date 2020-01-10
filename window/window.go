package window

import (
	"fmt"
	"time"

	"league.com/rulemaker/content"

	"github.com/gdamore/tcell"
	"league.com/rulemaker/diagnostics"
	"league.com/rulemaker/meta"
	"league.com/rulemaker/tokenizer"
	"league.com/rulemaker/view"
)

type Window interface {
	Run()
}

func NewWindow(c content.Content, metainfo meta.Meta, inputs, operations diagnostics.Set) (Window, error) {
	screen, e := tcell.NewScreen()
	if e != nil {
		return nil, e
	}
	if e := screen.Init(); e != nil {
		return nil, e
	}
	screen.SetStyle(defStyle)
	screen.EnableMouse()
	screen.Clear()

	w := &window{
		content:    c,
		metainfo:   metainfo,
		inputs:     inputs,
		operations: operations,
		screen:     screen,
	}
	w.resize()
	w.draw()
	return w, nil
}

var defStyle tcell.Style

// var mainBackground = tcell.NewRGBColor(0, 0, 63)
// var mainBackground = tcell.ColorBlack
// var mainStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(mainBackground)
// var lineNumberStyle = mainStyle.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver)
// var menuStyle tcell.Style = defStyle.Background(tcell.ColorSilver)

var mainStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(0, 0, 63))
var lineNumberStyle = mainStyle.Foreground(tcell.ColorSilver).Background(tcell.ColorBlack)
var lineNumberStyleCurrent = mainStyle.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver)
var menuStyle tcell.Style = defStyle.Background(tcell.ColorSilver)

type window struct {
	content content.Content

	screen        tcell.Screen
	width, height int
	vSplit        int

	titleView      view.View
	menuView       view.View
	statusView     view.View
	mainView       view.View
	lineNumberView view.View
	diagnosticView view.View

	metainfo   meta.Meta
	inputs     diagnostics.Set
	operations diagnostics.Set

	tokens      tokenizer.Tokens
	diagnostics []diagnostics.Diagnostic

	diagnosticViewPointers []point

	lineOffset   int
	columnOffset int
	cursorX      int
	cursorY      int
}

type point struct {
	x, y int
}

func (w *window) resize() {
	w.width, w.height = w.screen.Size()
	w.vSplit = w.width - 64
	if w.width < 128 {
		w.vSplit = w.width / 2
	}
	lineNumberViewWidth := 2
	l := len(w.content.Runes())
	for l > 0 {
		lineNumberViewWidth++
		l /= 10
	}

	w.titleView = view.NewView(w.screen, 0, w.width, 0, 1)
	w.menuView = view.NewView(w.screen, 0, w.width, 1, 2)
	w.lineNumberView = view.NewView(w.screen, 0, lineNumberViewWidth, 2, w.height-1)
	w.mainView = view.NewView(w.screen, lineNumberViewWidth, w.vSplit, 2, w.height-1)
	w.diagnosticView = view.NewView(w.screen, w.vSplit+1, w.width, 2, w.height-1)
	w.statusView = view.NewView(w.screen, 0, w.width, w.height-1, w.height)
}

func (w *window) clear() {
	w.resize()
	w.titleView.Clear(defStyle)
	w.menuView.Clear(menuStyle)
	w.lineNumberView.Clear(lineNumberStyle)
	w.mainView.Clear(mainStyle)
	w.diagnosticView.Clear(mainStyle)
	w.statusView.Clear(menuStyle)

	for row := 2; row < w.height-1; row++ {
		w.screen.SetContent(w.vSplit, row, tcell.RuneVLine, nil, mainStyle)
	}

	w.titleView.SetText("Rule Maker", 1, 0, defStyle.Bold(true))
	w.titleView.SetText(time.Now().Format("2006-01-02"), w.width-11, 0, defStyle.Bold(true))
	w.menuView.SetText("(Ctrl-Q) Quit  (Ctrl-N) Next Error  (Ctrl-P) Previous error", 1, 0, menuStyle)
	w.statusView.SetText(fmt.Sprintf("%s", w.content.Path()), 1, 0, menuStyle)
}

func (w *window) draw() {
	w.clear()
	w.mainView.SetOffsets(w.columnOffset, w.lineOffset)
	w.lineNumberView.SetOffsets(w.columnOffset, w.lineOffset)
	w.mainView.ShowCursor(w.cursorX, w.cursorY)

	w.tokens = tokenizer.Tokenize(w.content.Runes())
	w.diagnostics = diagnostics.ScanTokens(w.metainfo, w.inputs, w.operations, w.tokens)
	w.showText()
	w.showLineNumbers()
	w.showDiagnostics()
	w.showStatus()
	w.screen.Show()
}

func (w *window) ensureCursorVisible() {
	if w.cursorX < w.columnOffset {
		w.columnOffset = w.cursorX
	}
	if w.cursorX >= w.columnOffset+w.mainView.Width() {
		w.columnOffset = w.cursorX - w.mainView.Width() + 1
	}
	if w.cursorY < w.lineOffset {
		w.lineOffset = w.cursorY
	}
	if w.cursorY >= w.lineOffset+w.mainView.Height() {
		w.lineOffset = w.cursorY - w.mainView.Height() + 1
	}
}

func (w *window) showText() {
	diagnosticsIndex := 0
	for _, token := range w.tokens {
		var d *diagnostics.Diagnostic
		if diagnosticsIndex < len(w.diagnostics) &&
			w.diagnostics[diagnosticsIndex].Line == token.Line &&
			w.diagnostics[diagnosticsIndex].Column == token.Column {
			d = &w.diagnostics[diagnosticsIndex]
			diagnosticsIndex++
		}
		w.showToken(token, d)
	}
}

func (w *window) showToken(token tokenizer.Token, diagnosticMessage *diagnostics.Diagnostic) {
	if token.Line < w.lineOffset || token.Line >= w.lineOffset+w.mainView.Height() {
		return
	}
	tokenStyle := mainStyle
	switch token.Type {
	case tokenizer.CanonicalField:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite).Bold(true)
	case tokenizer.Variable, tokenizer.Operation:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite)
	case tokenizer.Label:
		tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0x8fffff))
	case tokenizer.Input:
		tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0x8fff8f))
	case tokenizer.OpenParen,
		tokenizer.CloseParen,
		tokenizer.EqualSign,
		tokenizer.Semicolon:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite).Bold(true)
	case tokenizer.Comment:
		// tokenStyle = tokenStyle.Foreground(tcell.ColorGray)
		tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0xadadad))
	case tokenizer.StringLiteral,
		tokenizer.IntegerLiteral,
		tokenizer.RealLiteral,
		tokenizer.BooleanLiteral,
		tokenizer.NilLiteral,
		tokenizer.DateLiteral,
		tokenizer.YearSpanLiteral,
		tokenizer.MonthSpanLiteral,
		tokenizer.DaySpanLiteral,
		tokenizer.TodayLiteral:
		// tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0xf0e68c))
		tokenStyle = tokenStyle.Foreground(tcell.ColorGold)
	case tokenizer.InvalidToken:
		// tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0xf0e68c))
		tokenStyle = tokenStyle.Foreground(tcell.ColorRed).Bold(true).Bold(true)
	}
	if diagnosticMessage != nil {
		tokenStyle = tokenStyle.Foreground(tcell.ColorRed).Bold(true).Bold(true)
	}

	w.mainView.SetText(token.Text, token.Column, token.Line, tokenStyle)
}

func (w *window) showLineNumbers() {
	format := fmt.Sprintf(" %%%dd ", w.lineNumberView.Width()-2)
	for i := 0; i < len(w.content.Runes()); i++ {
		number := fmt.Sprintf(format, i+1)
		if i == w.cursorY {
			w.lineNumberView.SetText(number, 0, i, lineNumberStyleCurrent)
		} else {
			w.lineNumberView.SetText(number, 0, i, lineNumberStyle)
		}
	}
}

func (w *window) showDiagnostics() {
	reportLine := 0
	w.diagnosticViewPointers = []point{}
	for _, d := range w.diagnostics {
		message := fmt.Sprintf("%d:%d %s", d.Line+1, d.Column+1, d.Message)
		lines := wrapLines(message, w.diagnosticView.Width())
		for _, line := range lines {
			w.diagnosticView.SetText(line, 0, reportLine, mainStyle)
			w.diagnosticViewPointers = append(w.diagnosticViewPointers, point{d.Column, d.Line})
			reportLine++
		}
	}
}

func wrapLines(str string, w int) (result []string) {
	if len(str) <= w {
		return []string{str}
	}
	result = []string{str[:w]}
	str = str[w:]
	for len(str) > w-4 {
		result = append(result, "    "+str[:w-4])
		str = str[w-4:]
	}
	result = append(result, "    "+str)
	return result
}

func (w *window) showStatus() {
	lineIndex := w.cursorY + w.lineOffset + 1
	column := w.cursorX + w.columnOffset + 1

	w.statusView.SetText(fmt.Sprintf("%s %d:%d", w.content.Path(), lineIndex, column), 1, 0, menuStyle)
}

func (w *window) Run() {
	for w.handleEvent(w.screen.PollEvent()) {
		w.draw()
	}
}

func (w *window) handleEvent(ev tcell.Event) bool {
	if w.cursorX < 0 || w.cursorY < 0 {
		w.screen.Fini()
		return false
	}
	switch ev := ev.(type) {
	case *tcell.EventResize:
		w.screen.Sync()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyRune {
			w.content.InsertRune(w.cursorY, w.cursorX, ev.Rune())
			w.cursorX++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			if w.cursorX > 0 {
				w.cursorX--
				w.content.RemoveRunes(w.cursorY, w.cursorX, 1)
			} else if w.cursorY > 0 {
				w.cursorX = w.content.JoinLines(w.cursorY)
				w.cursorY--
			}
		} else if ev.Key() == tcell.KeyDelete {
			w.content.RemoveRunes(w.cursorY, w.cursorX, 1)
		} else if ev.Key() == tcell.KeyEnter {
			w.content.SplitLine(w.cursorY, w.cursorX)
			w.cursorX = 0
			w.cursorY++
		} else if ev.Key() == tcell.KeyLeft && w.cursorX > 0 {
			w.cursorX--
		} else if ev.Key() == tcell.KeyRight {
			w.cursorX++
		} else if ev.Key() == tcell.KeyUp {
			if w.cursorY > 0 {
				w.cursorY--
			}
		} else if ev.Key() == tcell.KeyDown {
			w.cursorY++
		} else if ev.Key() == tcell.KeyHome {
			w.cursorX = 0
		} else if ev.Key() == tcell.KeyEnd {
			if w.cursorY < len(w.content.Runes()) {
				w.cursorX = len(w.content.Runes()[w.cursorY])
			}
		} else if ev.Key() == tcell.KeyPgUp {
			w.cursorY -= w.mainView.Height()
			if w.cursorY <= 0 {
				w.cursorY = 0
			}
			w.lineOffset -= w.mainView.Height()
			if w.lineOffset <= 0 {
				w.lineOffset = 0
			}
		} else if ev.Key() == tcell.KeyPgDn {
			lineNum := len(w.content.Runes())
			w.cursorY += w.mainView.Height()
			if w.cursorY >= lineNum {
				w.cursorY = lineNum - 1
			}
			w.lineOffset += w.mainView.Height()
			if w.lineOffset >= lineNum {
				w.lineOffset = lineNum - 1
			}
		} else if ev.Key() == tcell.KeyCtrlP {
			for i := len(w.diagnostics) - 1; i >= 0; i-- {
				d := w.diagnostics[i]
				if d.Line > w.cursorY {
					continue
				}
				if d.Line < w.cursorY || d.Column < w.cursorX {
					w.cursorX = d.Column
					w.cursorY = d.Line
					break
				}
			}
		} else if ev.Key() == tcell.KeyCtrlN {
			for _, d := range w.diagnostics {
				if d.Line < w.cursorY {
					continue
				}
				if d.Line > w.cursorY || d.Column > w.cursorX {
					w.cursorX = d.Column
					w.cursorY = d.Line
					break
				}
			}
		} else if ev.Key() == tcell.KeyCtrlQ {
			w.screen.Fini()
			return false
		}
		w.ensureCursorVisible()
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button == tcell.Button1 {
			if w.mainView.Contains(x, y) {
				w.cursorX, w.cursorY = w.mainView.CursorFromPhysicalCoordinates(x, y)
			} else if w.lineNumberView.Contains(x, y) {
				_, w.cursorY = w.lineNumberView.CursorFromPhysicalCoordinates(x, y)
				w.cursorX = 0
				w.ensureCursorVisible()
			} else if w.diagnosticView.Contains(x, y) {
				_, lineNum := w.diagnosticView.CursorFromPhysicalCoordinates(x, y)
				p := w.diagnosticViewPointers[lineNum]
				w.cursorX = p.x
				w.cursorY = p.y
			}
			w.ensureCursorVisible()
		}

		if button&tcell.WheelUp != 0 && w.lineOffset > 0 {
			w.lineOffset--
		}
		if button&tcell.WheelDown != 0 && w.lineOffset < len(w.tokens)-1 {
			w.lineOffset++
		}
	}
	return true
}
