package window

import (
	"fmt"
	"log"
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

	lineOffset   int
	columnOffset int
	cursorX      int
	cursorY      int
}

func (w *window) resize() {
	w.screen.Sync()
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
	w.menuView.SetText("(Ctrl-Q) Quit", 1, 0, menuStyle)
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
	log.Printf("### setCursorVisibility: %d:%d\n", w.cursorX, w.cursorY)
	if w.cursorX < w.columnOffset {
		w.columnOffset = w.cursorX
		log.Printf("### setCursorVisibility: columnOffset=%v\n", w.columnOffset)
	}
	if w.cursorX >= w.columnOffset+w.mainView.Width() {
		w.columnOffset = w.cursorX - w.mainView.Width() + 1
		log.Printf("### setCursorVisibility: columnOffset=%v\n", w.columnOffset)
	}
	if w.cursorY < w.lineOffset {
		w.lineOffset = w.cursorY
		log.Printf("### setCursorVisibility: lineOffset=%v\n", w.lineOffset)
	}
	if w.cursorY >= w.lineOffset+w.mainView.Height() {
		w.lineOffset = w.cursorY - w.mainView.Height() + 1
		log.Printf("### setCursorVisibility: lineOffset=%v\n", w.lineOffset)
	}
}

func (w *window) showText() {
	diagnosticsIndex := 0
	for lineIndex, line := range w.tokens {
		for tokenIndex, token := range line {
			var d *diagnostics.Diagnostic
			if diagnosticsIndex < len(w.diagnostics) &&
				w.diagnostics[diagnosticsIndex].LineIndex == lineIndex &&
				w.diagnostics[diagnosticsIndex].TokenIndex == tokenIndex {
				d = &w.diagnostics[diagnosticsIndex]
				diagnosticsIndex++
			}
			w.showToken(token, lineIndex, d)
		}
	}
}

func (w *window) showToken(token tokenizer.Token, tokenLine int, diagnosticMessage *diagnostics.Diagnostic) {
	if tokenLine < w.lineOffset || tokenLine >= w.lineOffset+w.mainView.Height() {
		return
	}
	tokenStyle := mainStyle
	switch token.Type {
	case tokenizer.Identifier:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite).Bold(true)
	case tokenizer.Variable:
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

	w.mainView.SetText(token.Text, token.Column, tokenLine, tokenStyle)
}

func (w *window) showLineNumbers() {
	format := fmt.Sprintf("%%%dd", w.lineNumberView.Width()-2)
	for i := 0; i < len(w.tokens); i++ {
		number := fmt.Sprintf(format, i+1)
		w.lineNumberView.SetText(number, 1, i, lineNumberStyle)
	}
}

func (w *window) showDiagnostics() {
	reportLine := 0
	for _, d := range w.diagnostics {
		token := w.tokens[d.LineIndex][d.TokenIndex]
		message := fmt.Sprintf("%d:%d %s", d.LineIndex+1, token.Column+1, d.Message)
		lines := wrapLines(message, w.diagnosticView.Width())
		for _, line := range lines {
			w.diagnosticView.SetText(line, 0, reportLine, mainStyle)
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
	switch ev := ev.(type) {
	case *tcell.EventResize:
		w.resize()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyCtrlQ {
			w.screen.Fini()
			return false
		} else if ev.Key() == tcell.KeyLeft && w.cursorX > 0 {
			w.cursorX--
		} else if ev.Key() == tcell.KeyRight {
			w.cursorX++
		} else if ev.Key() == tcell.KeyUp {
			if w.cursorY > 0 || w.lineOffset > 0 {
				w.cursorY--
			}
		} else if ev.Key() == tcell.KeyDown {
			w.cursorY++
		} else if ev.Key() == tcell.KeyCtrlA {
			w.cursorX = 0
		} else if ev.Key() == tcell.KeyCtrlE {
			lineIndex := w.cursorY
			lineLength := len(w.content.Runes()[lineIndex])
			w.cursorX = lineLength
			log.Printf("### handleEvent: w.cursorY=%d  w.lineOffset=%d  lineIndex=%d  lineLength=%d  w.cursorX=%d\n", w.cursorY, w.lineOffset, lineIndex, lineLength, w.cursorX)
		}
		w.ensureCursorVisible()
	case *tcell.EventMouse:
		x, y := ev.Position()
		button := ev.Buttons()
		if button == tcell.Button1 {
			log.Printf("### handleEvent: position=%d:%d\n", x, y)
			if w.mainView.Contains(x, y) {
				w.cursorX, w.cursorY = w.mainView.CursorFromPhysicalCoordinates(x, y)
			}
			if w.lineNumberView.Contains(x, y) {
				_, w.cursorY = w.mainView.CursorFromPhysicalCoordinates(x, y)
				w.cursorX = 0
				w.ensureCursorVisible()
			}
		}

		// log.Printf("### event: %d:%d %x\n", x, y, button)
		if button&tcell.WheelUp != 0 && w.lineOffset > 0 {
			w.lineOffset--
			w.cursorY++
			log.Printf("### handleEvent: lineOffset=%d\n", w.lineOffset)
		}
		if button&tcell.WheelDown != 0 && w.lineOffset < len(w.tokens)-1 {
			w.lineOffset++
			w.cursorY--
			log.Printf("### handleEvent: lineOffset=%d\n", w.lineOffset)
		}
	}
	return true
}
