package window

import (
	"fmt"
	"log"
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

func NewWindow(c content.Content, metainfo meta.Meta, inputs, operations model.Set, theme style.Theme) (Window, error) {
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
	content content.Content

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
	w.setText(w.statusView, fmt.Sprintf("%s", w.content.Path()), 0, 1, menuStyle)
}

func (w *window) draw() {
	w.clear()
	w.tokens = tokenizer.TokenizeRunes(w.content.Runes())
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

	w.ShowCursor(w.mainView)
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
	w.lineNumberView.SetCursor(0, w.mainView.CursorLine)
	w.lineNumberView.LineOffset = w.mainView.LineOffset

	for i := 0; i < w.mainView.TotalLines; i++ {
		number := fmt.Sprintf(format, i+1)
		if i == w.mainView.CursorLine {
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
			w.diagnosticsViewPointers = append(w.diagnosticsViewPointers, point{d.Token.Column, d.Token.Line})
			reportLine++
		}
	}
}

func (w *window) showCompletions() {
	complitions := w.parser.Completions(w.mainView.CursorLine, w.mainView.CursorColumn)
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
	w.setText(w.statusView, fmt.Sprintf("%s %d:%d", w.content.Path(), w.mainView.CursorLine+1, w.mainView.CursorColumn+1), 0, 1, menuStyle)
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
	column, line := w.mainView.CursorColumn, w.mainView.CursorLine
	switch ev := ev.(type) {
	case *tcell.EventResize:
		w.screen.Sync()
	case *tcell.EventKey:
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
				if d.Token.Line > line {
					continue
				}
				if d.Token.Line < line || d.Token.Column < column {
					w.mainView.SetCursor(d.Token.Column, d.Token.Line)
					break
				}
			}
		} else if ev.Key() == tcell.KeyCtrlN {
			for _, d := range w.parser.Diagnostics() {
				if d.Token.Line < line {
					continue
				}
				if d.Token.Line > line || d.Token.Column > column {
					w.mainView.SetCursor(d.Token.Column, d.Token.Line)
					break
				}
			}
		} else if ev.Key() == tcell.KeyTab {
			text := w.parser.Completion(0)
			w.content.InsertRunes(line, column, []rune(text))
			w.mainView.MoveCursor(len(text), 0)
		} else if ev.Key() == tcell.KeyCtrlQ {
			w.screen.Fini()
			return false
		}
		w.ShowCursor(w.mainView)
		w.completionsView.LineOffset = 0
		log.Printf("### key: name: %v   key: %v   mod: %v   rune: %v\n", ev.Name(), ev.Key(), ev.Modifiers(), ev.Rune())
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
				w.content.InsertRunes(line, column, []rune(text))
				w.mainView.MoveCursor(len(text), 0)
			}
			w.ShowCursor(w.mainView)
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
			if w.mainView.Contains(x, y) || w.lineNumberView.Contains(x, y) {
				w.mainView.Scroll(lines)
			} else if w.completionsView.Contains(x, y) {
				w.completionsView.Scroll(lines)
			} else if w.diagnosticsView.Contains(x, y) {
				w.diagnosticsView.Scroll(lines)
			}
			w.ShowCursor(w.mainView)
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
	v.TotalLines = 0
}

func (w *window) ShowCursor(v *view.View) {
	if v.CursorColumn < v.ColumnOffset || v.CursorLine < v.LineOffset || v.CursorColumn >= v.ColumnOffset+v.Width || v.CursorLine >= v.LineOffset+v.Height {
		w.screen.HideCursor()
		return
	}
	w.screen.ShowCursor(v.CursorColumn-v.ColumnOffset+v.Left, v.CursorLine-v.LineOffset+v.Top)
}
