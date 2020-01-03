package window

import (
	"fmt"
	"time"

	"league.com/rulemaker/content"

	"github.com/gdamore/tcell"
	"league.com/rulemaker/meta"
	"league.com/rulemaker/parser"
	"league.com/rulemaker/tokenizer"
	"league.com/rulemaker/view"
)

type Window interface {
	Run()
}

func NewWindow(c content.Content, metainfo meta.Meta, inputs, operations parser.Set) (Window, error) {
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
	inputs     parser.Set
	operations parser.Set
	lineOffset int
}

func (w *window) resize() {
	w.width, w.height = w.screen.Size()
	w.vSplit = w.width - 64
	if w.width < 128 {
		w.vSplit = w.width / 2
	}

	w.titleView = view.NewView(w.screen, 0, w.width, 0, 1)
	w.menuView = view.NewView(w.screen, 0, w.width, 1, 2)
	w.lineNumberView = view.NewView(w.screen, 0, 5, 2, w.height-1)
	w.mainView = view.NewView(w.screen, 5, w.vSplit, 2, w.height-1)
	w.diagnosticView = view.NewView(w.screen, w.vSplit, w.width, 2, w.height-1)
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
	w.mainView.SetOffsets(0, w.lineOffset)
	w.lineNumberView.SetOffsets(0, w.lineOffset)

	tokenizer.Tokenize(w.content.Runes(),
		parser.NewParser(w.metainfo, w.inputs, w.operations,
			newTee(
				&errors{view: w.diagnosticView},
				&tokens{view: w.mainView, lineOffset: w.lineOffset},
				&lineNumbers{view: w.lineNumberView, lineOffset: w.lineOffset},
			),
		),
	)

	w.screen.Show()
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
		// log.Printf("Key=%v\n", ev.Key())
		// log.Printf("Rune=%v\n", ev.Rune())
		if ev.Key() == tcell.KeyCtrlQ {
			w.screen.Fini()
			return false
		}
	case *tcell.EventMouse:
		// x, y := ev.Position()
		button := ev.Buttons()
		// log.Printf("### event: %d:%d %x\n", x, y, button)
		if button&tcell.WheelUp != 0 && w.lineOffset > 0 {
			w.lineOffset--
		}
		if button&tcell.WheelDown != 0 && w.lineOffset < len(w.content.Runes())-2 {
			w.lineOffset++
		}
	}
	return true
}

type tee struct {
	outStreams []parser.Tokens
}

func newTee(outStreams ...parser.Tokens) parser.Tokens {
	return &tee{
		outStreams: outStreams,
	}
}

func (t *tee) Token(token parser.ParsedToken) {
	for _, out := range t.outStreams {
		out.Token(token)
	}
}

func (t *tee) Done() {
	for _, out := range t.outStreams {
		out.Done()
	}
}

type tokens struct {
	view       view.View
	lineOffset int
}

func (t *tokens) Token(token parser.ParsedToken) {
	if token.Line < t.lineOffset || token.Line >= t.lineOffset+t.view.Height() {
		return
	}
	tokenStyle := mainStyle
	switch token.Type {
	case tokenizer.CanonicalField, tokenizer.Function:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite).Bold(true)
	case tokenizer.Variable:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite)
	case tokenizer.Label:
		tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0x8fffff))
	case tokenizer.Input:
		tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0x8fff8f))
	case tokenizer.OpenParen, tokenizer.CloseParen, tokenizer.EqualSign, tokenizer.Semicolon:
		tokenStyle = tokenStyle.Foreground(tcell.ColorWhite).Bold(true)
	case tokenizer.Comment:
		// tokenStyle = tokenStyle.Foreground(tcell.ColorGray)
		tokenStyle = tokenStyle.Foreground(tcell.NewHexColor(0xadadad))
	case tokenizer.StringLiteral,
		tokenizer.IntegerLiteral,
		tokenizer.FloatingPointLiteral,
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
	if token.Diagnostic != "" {
		tokenStyle = tokenStyle.Foreground(tcell.ColorRed).Bold(true).Bold(true)
	}

	t.view.SetText(token.Text, token.StartColumn, token.Line, tokenStyle)
}

func (t *tokens) Done() {}

type lineNumbers struct {
	view       view.View
	lineOffset int
}

func (n *lineNumbers) Token(token parser.ParsedToken) {
	number := fmt.Sprintf("%4d", token.Line+1)
	n.view.SetText(number, 0, token.Line, lineNumberStyle)
}

func (n *lineNumbers) Done() {}

type errors struct {
	view       view.View
	reportLine int
}

func (t *errors) Token(token parser.ParsedToken) {
	if token.Diagnostic != "" {
		message := fmt.Sprintf("%d:%d %s", token.Line+1, token.StartColumn+1, token.Diagnostic)
		lines := wrapLines(message, t.view.Width())
		for _, line := range lines {
			t.view.SetText(line, 1, t.reportLine, mainStyle)
			t.reportLine++
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

func (t *errors) Done() {

}
