package window

import (
	"fmt"
	"time"

	"league.com/rulemaker/content"
	"league.com/rulemaker/model"
	"league.com/rulemaker/parser"

	"github.com/gdamore/tcell"
	"league.com/rulemaker/meta"
	"league.com/rulemaker/tokenizer"
	"league.com/rulemaker/view"
)

type Window interface {
	Run()
}

func NewWindow(c content.Content, metainfo meta.Meta, inputs, operations model.Set) (Window, error) {
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
		content: c,
		parser:  parser.NewParser(metainfo, inputs, operations),
		screen:  screen,
	}

	w.titleView = view.NewView(w.screen, defStyle)
	w.menuView = view.NewView(w.screen, menuStyle)
	w.lineNumberView = view.NewView(w.screen, lineNumberStyle)
	w.mainView = view.NewView(w.screen, mainStyle)
	w.diagnosticsView = view.NewView(w.screen, mainStyle)
	w.completionsView = view.NewView(w.screen, mainStyle)
	w.statusView = view.NewView(w.screen, menuStyle)

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

	parser parser.Parser

	screen         tcell.Screen
	width, height  int
	vSplit, hSplit int

	titleView       view.View
	menuView        view.View
	mainView        view.View
	lineNumberView  view.View
	diagnosticsView view.View
	completionsView view.View
	statusView      view.View

	tokens                  tokenizer.Tokens
	diagnosticsViewPointers []point
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
	w.hSplit = w.height / 2
	lineNumberViewWidth := 2
	l := len(w.content.Runes())
	for l > 0 {
		lineNumberViewWidth++
		l /= 10
	}

	w.titleView.Resize(0, w.width, 0, 1)
	w.menuView.Resize(0, w.width, 1, 2)
	w.lineNumberView.Resize(0, lineNumberViewWidth, 2, w.height-1)
	w.mainView.Resize(lineNumberViewWidth, w.vSplit, 2, w.height-1)
	w.diagnosticsView.Resize(w.vSplit+1, w.width, w.hSplit+1, w.height-1)
	w.completionsView.Resize(w.vSplit+1, w.width, 2, w.hSplit)
	w.statusView.Resize(0, w.width, w.height-1, w.height)
}

func (w *window) clear() {
	w.resize()
	w.titleView.Clear()
	w.menuView.Clear()
	w.lineNumberView.Clear()
	w.mainView.Clear()
	w.diagnosticsView.Clear()
	w.completionsView.Clear()
	w.statusView.Clear()

	for row := 2; row < w.height-1; row++ {
		w.screen.SetContent(w.vSplit, row, tcell.RuneVLine, nil, mainStyle)
	}

	w.screen.SetContent(w.vSplit, w.hSplit, tcell.RuneLTee, nil, mainStyle)
	for col := w.vSplit + 1; col < w.width; col++ {
		w.screen.SetContent(col, w.hSplit, tcell.RuneHLine, nil, mainStyle)
	}

	w.titleView.SetText("Rule Maker", 1, 0, defStyle.Bold(true))
	w.titleView.SetText(time.Now().Format("2006-01-02"), w.width-11, 0, defStyle.Bold(true))
	w.menuView.SetText("(Ctrl-Q) Quit  (Ctrl-N) Next Error  (Ctrl-P) Previous error", 1, 0, menuStyle)
	w.statusView.SetText(fmt.Sprintf("%s", w.content.Path()), 1, 0, menuStyle)
}

func (w *window) draw() {
	w.clear()
	w.tokens = tokenizer.Tokenize(w.content.Runes())
	w.parser.Parse(w.tokens)
	w.showText()
	w.showLineNumbers()
	w.showDiagnostics()
	w.showCompletions()
	w.showStatus()
	w.screen.Show()
}

func (w *window) showText() {
	diagnosticsIndex := 0
	diagnostics := w.parser.Diagnostics()
	for _, token := range w.tokens {
		var d *parser.Diagnostic
		if diagnosticsIndex < len(diagnostics) &&
			diagnostics[diagnosticsIndex].Line == token.Line &&
			diagnostics[diagnosticsIndex].Column == token.Column {
			d = &diagnostics[diagnosticsIndex]
			diagnosticsIndex++
		}
		w.showToken(token, d)
	}
	w.mainView.ShowCursor()
}

func (w *window) showToken(token tokenizer.Token, diagnosticMessage *parser.Diagnostic) {
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
	_, lineNum := w.mainView.Cursor()
	w.lineNumberView.SetCursor(0, lineNum)
	_, lineOffset := w.mainView.Offsets()
	w.lineNumberView.SetOffsets(0, lineOffset)

	for i := 0; i < w.mainView.TotalLines(); i++ {
		number := fmt.Sprintf(format, i+1)
		if i == lineNum {
			w.lineNumberView.SetText(number, 0, i, lineNumberStyleCurrent)
		} else {
			w.lineNumberView.SetText(number, 0, i, lineNumberStyle)
		}
	}
}

func (w *window) showDiagnostics() {
	reportLine := 0
	w.diagnosticsViewPointers = []point{}
	for _, d := range w.parser.Diagnostics() {
		message := fmt.Sprintf("%d:%d %s", d.Line+1, d.Column+1, d.Message)
		lines := wrapLines(message, w.diagnosticsView.Width())
		for _, line := range lines {
			w.diagnosticsView.SetText(line, 0, reportLine, mainStyle)
			w.diagnosticsViewPointers = append(w.diagnosticsViewPointers, point{d.Column, d.Line})
			reportLine++
		}
	}
}

func (w *window) showCompletions() {
	names := w.parser.Completions(w.mainView.Cursor())
	for i, name := range names {
		w.completionsView.SetText(name, 0, i, mainStyle)
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
	column, line := w.mainView.Cursor()
	w.statusView.SetText(fmt.Sprintf("%s %d:%d", w.content.Path(), line+1, column+1), 1, 0, menuStyle)
}

func (w *window) Run() {
	for w.handleEvent() {
		w.draw()
	}
}

func (w *window) handleEvent() bool {
	ev := w.screen.PollEvent()
	for {
		if ev, mouseEvent := ev.(*tcell.EventMouse); !mouseEvent || ev.Buttons() != 0 {
			break
		}
		ev = w.screen.PollEvent()
	}
	switch ev := ev.(type) {
	case *tcell.EventResize:
		w.screen.Sync()
	case *tcell.EventKey:
		column, line := w.mainView.Cursor()
		if ev.Key() == tcell.KeyRune {
			w.content.InsertRune(line, column, ev.Rune())
			w.mainView.SetCursor(column+1, line)
			if ev.Rune() == '(' {
				w.content.InsertRune(line, column+1, ')')
			}
			if ev.Rune() == '"' {
				w.content.InsertRune(line, column+1, '"')
			}
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			if column > 0 {
				w.mainView.SetCursor(column-1, line)
				w.content.RemoveRunes(line, column-1, 1)
			} else if line > 0 {
				column = w.content.JoinLines(line)
				w.mainView.SetCursor(column, line-1)
			}
		} else if ev.Key() == tcell.KeyDelete {
			w.content.RemoveRunes(line, column, 1)
		} else if ev.Key() == tcell.KeyEnter {
			w.content.SplitLine(line, column)
			w.mainView.SetCursor(0, line+1)
		} else if ev.Key() == tcell.KeyLeft {
			w.mainView.MoveCursor(-1, 0)
		} else if ev.Key() == tcell.KeyRight {
			w.mainView.MoveCursor(1, 0)
		} else if ev.Key() == tcell.KeyUp {
			w.mainView.MoveCursor(0, -1)
		} else if ev.Key() == tcell.KeyDown {
			w.mainView.MoveCursor(0, 1)
		} else if ev.Key() == tcell.KeyHome {
			w.mainView.SetCursor(0, line)
		} else if ev.Key() == tcell.KeyEnd {
			if line < len(w.content.Runes()) {
				w.mainView.SetCursor(len(w.content.Runes()[line]), line)
			}
		} else if ev.Key() == tcell.KeyPgUp {
			w.mainView.PageUp()
		} else if ev.Key() == tcell.KeyPgDn {
			w.mainView.PageDown()
		} else if ev.Key() == tcell.KeyCtrlP {
			for i := len(w.parser.Diagnostics()) - 1; i >= 0; i-- {
				d := w.parser.Diagnostics()[i]
				if d.Line > line {
					continue
				}
				if d.Line < line || d.Column < column {
					w.mainView.SetCursor(d.Column, d.Line)
					break
				}
			}
		} else if ev.Key() == tcell.KeyCtrlN {
			for _, d := range w.parser.Diagnostics() {
				if d.Line < line {
					continue
				}
				if d.Line > line || d.Column > column {
					w.mainView.SetCursor(d.Column, d.Line)
					break
				}
			}
		} else if ev.Key() == tcell.KeyTab {
			text := w.parser.Completion(0)
			col, line := w.mainView.Cursor()
			w.content.InsertRunes(line, col, []rune(text))
			w.mainView.MoveCursor(len(text), 0)
		} else if ev.Key() == tcell.KeyCtrlQ {
			w.screen.Fini()
			return false
		}
		w.mainView.ShowCursor()
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button == tcell.Button1 {
			if w.mainView.Contains(x, y) {
				w.mainView.SetCursor(w.mainView.CursorFromScreenCoordinates(x, y))
			} else if w.lineNumberView.Contains(x, y) {
				_, line := w.lineNumberView.CursorFromScreenCoordinates(x, y)
				w.mainView.SetCursor(0, line)
			} else if w.diagnosticsView.Contains(x, y) {
				_, lineNum := w.diagnosticsView.CursorFromScreenCoordinates(x, y)
				if lineNum < len(w.diagnosticsViewPointers) {
					p := w.diagnosticsViewPointers[lineNum]
					w.mainView.SetCursor(p.x, p.y)
				}
			} else if w.completionsView.Contains(x, y) {
				_, lineNum := w.completionsView.CursorFromScreenCoordinates(x, y)
				text := w.parser.Completion(lineNum)
				col, line := w.mainView.Cursor()
				w.content.InsertRunes(line, col, []rune(text))
				w.mainView.MoveCursor(len(text), 0)
			}
			w.mainView.ShowCursor()
		}

		lines := 0
		if button&tcell.WheelUp != 0 {
			lines = -1
		} else if button&tcell.WheelDown != 0 {
			lines = 1
		}

		if lines != 0 {
			if w.mainView.Contains(x, y) || w.lineNumberView.Contains(x, y) {
				w.mainView.Scroll(lines)
			} else if w.completionsView.Contains(x, y) {
				w.completionsView.Scroll(lines)
			} else if w.diagnosticsView.Contains(x, y) {
				w.diagnosticsView.Scroll(lines)
			}
			w.mainView.ShowCursor()
		}
	}
	return true
}
