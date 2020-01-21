package window

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"league.com/rulemaker/content"
	"league.com/rulemaker/meta"
	"league.com/rulemaker/model"
	"league.com/rulemaker/parser"
	"league.com/rulemaker/style"
	"league.com/rulemaker/tokenizer"
	"league.com/rulemaker/view"
)

type Window interface {
	Run()
}

func NewWindow(c *content.Content, metainfo meta.Meta, inputs, operations model.Set, theme style.Theme) (Window, error) {
	if theme == style.BlueTheme {
		mainStyle = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.Color17)
		lineNumberStyle = mainStyle.Foreground(tcell.ColorSilver).Background(tcell.Color235)
		lineNumberStyleCurrent = mainStyle.Foreground(tcell.Color231).Background(tcell.ColorGrey)
		menuStyle = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.ColorSilver)
	} else if theme == style.DarkTheme {
		mainStyle = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.Color235)
		lineNumberStyle = mainStyle.Foreground(tcell.Color231).Background(tcell.Color238)
		lineNumberStyleCurrent = mainStyle.Foreground(tcell.Color231).Background(tcell.Color242)
		menuStyle = tcell.StyleDefault.Foreground(tcell.Color231).Background(tcell.Color244)
	} else if theme == style.LightTheme {
		mainStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.Color231)
		lineNumberStyle = mainStyle.Foreground(tcell.ColorBlack).Background(tcell.Color250)
		lineNumberStyleCurrent = mainStyle.Foreground(tcell.ColorBlack).Background(tcell.Color248).Bold(true)
		menuStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorSilver)
	}

	screen, e := tcell.NewScreen()
	if e != nil {
		return nil, e
	}
	if e := screen.Init(); e != nil {
		return nil, e
	}
	screen.SetStyle(mainStyle)
	screen.EnableMouse()
	screen.Clear()

	w := &window{
		theme:   theme,
		content: c,
		parser:  parser.NewParser(metainfo, inputs, operations),
		screen:  screen,
	}

	w.titleView = view.NewView(mainStyle)
	w.menuView = view.NewView(menuStyle)
	w.lineNumberView = view.NewView(lineNumberStyle)
	w.mainView = view.NewView(mainStyle)
	w.diagnosticsView = view.NewView(mainStyle)
	w.completionsView = view.NewView(mainStyle)
	w.statusView = view.NewView(menuStyle)

	return w, nil
}

var (
	mainStyle              tcell.Style
	lineNumberStyle        tcell.Style
	lineNumberStyleCurrent tcell.Style
	menuStyle              tcell.Style
)

type window struct {
	theme   style.Theme
	content *content.Content

	parser *parser.Parser

	screen         tcell.Screen
	width, height  int
	vSplit, hSplit int

	titleView       *view.View
	menuView        *view.View
	mainView        *view.View
	lineNumberView  *view.View
	diagnosticsView *view.View
	completionsView *view.View
	statusView      *view.View

	tokens                  tokenizer.Tokens
	diagnosticsViewPointers []point
}

type point struct {
	line, column int
}

func (w *window) resize() {
	w.width, w.height = w.screen.Size()
	w.vSplit = w.width - 64
	if w.width < 128 {
		w.vSplit = w.width / 2
	}
	w.hSplit = w.height / 2
	lineNumberViewWidth := 2
	l := len(w.content.Runes)
	for l > 0 {
		lineNumberViewWidth++
		l /= 10
	}

	w.titleView.Resize(0, 1, 0, w.width)
	w.menuView.Resize(1, 2, 0, w.width)
	w.lineNumberView.Resize(2, w.height-1, 0, lineNumberViewWidth)
	w.mainView.Resize(2, w.height-1, lineNumberViewWidth, w.vSplit)
	w.diagnosticsView.Resize(w.hSplit+1, w.height-1, w.vSplit+1, w.width)
	w.completionsView.Resize(2, w.hSplit, w.vSplit+1, w.width)
	w.statusView.Resize(w.height-1, w.height, 0, w.width)
}

func (w *window) clear() {
	w.resize()
	w.Clear(w.titleView)
	w.Clear(w.menuView)
	w.Clear(w.lineNumberView)
	w.Clear(w.mainView)
	w.Clear(w.diagnosticsView)
	w.Clear(w.completionsView)
	w.Clear(w.statusView)

	for row := 2; row < w.height-1; row++ {
		w.screen.SetContent(w.vSplit, row, tcell.RuneVLine, nil, mainStyle)
	}

	w.screen.SetContent(w.vSplit, w.hSplit, tcell.RuneLTee, nil, mainStyle)
	for col := w.vSplit + 1; col < w.width; col++ {
		w.screen.SetContent(col, w.hSplit, tcell.RuneHLine, nil, mainStyle)
	}

	w.setText(w.titleView, "Rule Maker", 0, 0, mainStyle.Bold(true))
	w.setText(w.titleView, time.Now().Format("2006-01-02"), 0, w.width-11, mainStyle.Bold(true))
	w.setText(w.menuView, "(Ctrl-Q) Quit  (Ctrl-N) Next Error  (Ctrl-P) Previous error", 0, 1, menuStyle)
	w.setText(w.statusView, fmt.Sprintf("%s", w.content.Path), 0, 1, menuStyle)
}

func (w *window) draw() {
	w.clear()
	w.tokens = tokenizer.TokenizeRunes(w.content.Runes)
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
			diagnostics[diagnosticsIndex].Token.Line == token.Line &&
			diagnostics[diagnosticsIndex].Token.Column == token.Column {
			d = &diagnostics[diagnosticsIndex]
			diagnosticsIndex++
		}
		w.showToken(token, d)
	}

	w.ShowCursor()
}

func (w *window) showToken(token tokenizer.Token, diagnosticMessage *parser.Diagnostic) {
	tt := token.Type
	if diagnosticMessage != nil {
		tt = tokenizer.InvalidToken
	}

	w.setText(w.mainView, token.Text, token.Line, token.Column, style.TokenStyle(tt, w.theme))
}

func (w *window) setText(v *view.View, text string, line, column int, style tcell.Style) {
	text, line, column = v.ClipText(text, line, column)
	for i, ch := range text {
		w.screen.SetContent(column+i, line, ch, nil, style)
	}
}

func (w *window) showLineNumbers() {
	format := fmt.Sprintf(" %%%dd ", w.lineNumberView.Width-2)
	w.lineNumberView.LineOffset = w.mainView.LineOffset

	for i := 0; i < len(w.content.Runes); i++ {
		number := fmt.Sprintf(format, i+1)
		if i == w.content.Cursor.Line {
			w.setText(w.lineNumberView, number, i, 0, lineNumberStyleCurrent)
		} else {
			w.setText(w.lineNumberView, number, i, 0, lineNumberStyle)
		}
	}
}

func (w *window) showDiagnostics() {
	reportLine := 0
	w.diagnosticsViewPointers = []point{}
	for _, d := range w.parser.Diagnostics() {
		message := fmt.Sprintf("%d:%d %s", d.Token.Line+1, d.Token.Column+1, d.Message)
		lines := wrapLines(message, w.diagnosticsView.Width)
		for _, line := range lines {
			w.setText(w.diagnosticsView, line, reportLine, 0, mainStyle)
			w.diagnosticsViewPointers = append(w.diagnosticsViewPointers, point{d.Token.Line, d.Token.Column})
			reportLine++
		}
	}
}

func (w *window) showCompletions() {
	complitions := w.parser.Completions(w.content.Cursor.Line, w.content.Cursor.Column)
	for i, complition := range complitions {
		text := complition.Name
		if complition.TokenType == tokenizer.CanonicalField {
			text = " " + complition.Name
		}
		w.setText(w.completionsView, text, i, 0, style.TokenStyle(complition.TokenType, w.theme))
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
	w.setText(w.statusView, fmt.Sprintf("%s %d:%d", w.content.Path, w.content.Cursor.Line+1, w.content.Cursor.Column+1), 0, 1, menuStyle)
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
		if ev.Key() == tcell.KeyRune {
			w.content.InsertRune(ev.Rune())
			if ev.Rune() == '(' {
				w.content.InsertRune(')')
				w.content.MoveCursorLeft(1)
			}
			if ev.Rune() == '"' {
				w.content.InsertRune('"')
				w.content.MoveCursorLeft(1)
			}
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			w.content.DeleteLeft()
		} else if ev.Key() == tcell.KeyDelete {
			w.content.DeleteRight()
		} else if ev.Key() == tcell.KeyEnter {
			w.content.SplitLine()
		} else if ev.Key() == tcell.KeyLeft {
			w.content.MoveCursorLeft(1)
		} else if ev.Key() == tcell.KeyRight {
			w.content.MoveCursorRight(1)
		} else if ev.Key() == tcell.KeyUp {
			w.content.MoveCursorUp(1)
		} else if ev.Key() == tcell.KeyDown {
			w.content.MoveCursorDown(1, w.mainView.Height-1)
		} else if ev.Key() == tcell.KeyHome {
			w.content.MoveCursorToBol()
		} else if ev.Key() == tcell.KeyEnd {
			w.content.MoveCursorToEol()
		} else if ev.Key() == tcell.KeyPgUp {
			w.content.MoveCursorUp(w.mainView.Height)
		} else if ev.Key() == tcell.KeyPgDn {
			w.content.MoveCursorDown(w.mainView.Height, w.mainView.Height-1)
		} else if ev.Key() == tcell.KeyCtrlP {
			for i := len(w.parser.Diagnostics()) - 1; i >= 0; i-- {
				d := w.parser.Diagnostics()[i]
				if d.Token.Line > w.content.Cursor.Line {
					continue
				}
				if d.Token.Line < w.content.Cursor.Line || d.Token.Column < w.content.Cursor.Column {
					w.content.SetCursor(d.Token.Line, d.Token.Column)
					break
				}
			}
		} else if ev.Key() == tcell.KeyCtrlN {
			for _, d := range w.parser.Diagnostics() {
				if d.Token.Line < w.content.Cursor.Line {
					continue
				}
				if d.Token.Line > w.content.Cursor.Line || d.Token.Column > w.content.Cursor.Column {
					w.content.SetCursor(d.Token.Line, d.Token.Column)
					break
				}
			}
		} else if ev.Key() == tcell.KeyTab {
			text := w.parser.Completion(0)
			w.content.InsertRunes([]rune(text))
		} else if ev.Key() == tcell.KeyCtrlQ {
			w.screen.Fini()
			return false
		}
		w.mainView.MakeCursorVisible(w.content.Cursor.Line, w.content.Cursor.Column)
		w.ShowCursor()
		w.completionsView.LineOffset = 0
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button == tcell.Button1 {
			if w.mainView.Contains(y, x) {
				w.content.SetCursor(w.mainView.CursorFromScreenCoordinates(y, x))
			} else if w.lineNumberView.Contains(y, x) {
				line, _ := w.lineNumberView.CursorFromScreenCoordinates(y, x)
				w.content.SetCursor(line, 0)
			} else if w.diagnosticsView.Contains(y, x) {
				lineNum, _ := w.diagnosticsView.CursorFromScreenCoordinates(y, x)
				if lineNum < len(w.diagnosticsViewPointers) {
					p := w.diagnosticsViewPointers[lineNum]
					w.content.SetCursor(p.line, p.column)
				}
			} else if w.completionsView.Contains(y, x) {
				lineNum, _ := w.completionsView.CursorFromScreenCoordinates(y, x)
				text := w.parser.Completion(lineNum)
				w.content.InsertRunes([]rune(text))
			}
			w.mainView.MakeCursorVisible(w.content.Cursor.Line, w.content.Cursor.Column)
			w.ShowCursor()
			w.completionsView.LineOffset = 0
			return true
		}

		lines := 0
		if button&tcell.WheelUp != 0 {
			lines = -1
		} else if button&tcell.WheelDown != 0 {
			lines = 1
		}

		if lines != 0 {
			if w.mainView.Contains(y, x) || w.lineNumberView.Contains(y, x) {
				w.mainView.Scroll(lines, len(w.content.Runes)-1)
			} else if w.completionsView.Contains(y, x) {
				w.completionsView.Scroll(lines, w.parser.TotalCompletions()-1)
			} else if w.diagnosticsView.Contains(y, x) {
				w.diagnosticsView.Scroll(lines, len(w.diagnosticsViewPointers)-1)
			}
			w.ShowCursor()
		}
	}
	return true
}

func (w *window) Clear(v *view.View) {
	for row := 0; row < v.Height; row++ {
		for col := 0; col < v.Width; col++ {
			w.screen.SetContent(col+v.Left, row+v.Top, ' ', nil, v.Style)
		}
	}
}

func (w *window) ShowCursor() {
	if w.content.Cursor.Column < w.mainView.ColumnOffset ||
		w.content.Cursor.Line < w.mainView.LineOffset ||
		w.content.Cursor.Column >= w.mainView.ColumnOffset+w.mainView.Width ||
		w.content.Cursor.Line >= w.mainView.LineOffset+w.mainView.Height {
		w.screen.HideCursor()
		return
	}
	w.screen.ShowCursor(w.content.Cursor.Column-w.mainView.ColumnOffset+w.mainView.Left, w.content.Cursor.Line-w.mainView.LineOffset+w.mainView.Top)
}
