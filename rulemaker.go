package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"league.com/rulemaker/parser"
	"league.com/rulemaker/tokenizer"

	"github.com/gdamore/tcell"
)

func main() {
	file, err := os.Open("emp.rules")
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.Clear()

	w := window{screen: s, content: string(content)}

	for {
		w.draw()
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			// log.Printf("Key=%v\n", ev.Key())
			// log.Printf("Rune=%v\n", ev.Rune())
			if ev.Key() == tcell.KeyCtrlL {
				s.Sync()
			} else if ev.Key() == tcell.KeyCtrlQ {
				s.Fini()
				os.Exit(0)
			} else {
				if ev.Rune() == 'C' || ev.Rune() == 'c' {
					s.SetContent(0, 0, ' ', nil, defStyle)
				}
			}
		case *tcell.EventMouse:
			button := ev.Buttons()
			if button&tcell.WheelUp != 0 && lineOffset > 0 {
				lineOffset--
			}
			if button&tcell.WheelDown != 0 {
				lineOffset++
			}
		}
	}
}

type window struct {
	screen  tcell.Screen
	title   string
	status  string
	content string
}

var lineOffset = 0

var mainStyle = tcell.StyleDefault.Background(colorDeepBlue).Foreground(tcell.ColorWhite)

func (w *window) draw() {
	width, height := w.screen.Size()
	var vSplit = 64

	for row := 2; row < height-1; row++ {
		for col := 0; col < width; col++ {
			w.screen.SetContent(col, row, ' ', nil, mainStyle)
		}
	}

	// w.showPalette()

	for row := 1; row < height-1; row++ {
		w.screen.SetContent(vSplit, row, tcell.RuneVLine, nil, mainStyle)
	}

	for col := 0; col < width; col++ {
		w.screen.SetContent(col, 0, ' ', nil, defStyle)
		w.screen.SetContent(col, 1, ' ', nil, menuStyle)
		w.screen.SetContent(col, height-1, ' ', nil, menuStyle)
	}
	emitStr(w.screen, 1, 0, defStyle.Bold(true), "Rule Maker")
	emitStr(w.screen, width-11, 0, defStyle.Bold(true), time.Now().Format("2006-01-02"))

	emitStr(w.screen, 1, 1, menuStyle, "(Ctrl-Q) Quit")
	emitStr(w.screen, 1, height-1, menuStyle, fmt.Sprintf("emp.rules:%d", lineOffset+1))

	t := newTee(
		&tokens{w: w, vSplit: vSplit, height: height},
		&errors{w: w, vSplit: vSplit, height: height},
	)

	parser.Parse(splitLines(w.content), nil, nil, t)

	w.screen.Show()
}

type tokens struct {
	w              *window
	vSplit, height int
}

func (t *tokens) Token(token parser.ParsedToken) {
	if token.Line < lineOffset || token.Line >= lineOffset+t.height-3 {
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
	emitStr(t.w.screen, t.vSplit+token.StartColumn+1, token.Line+2-lineOffset, tokenStyle, token.Text)
}

func (t *tokens) Done() {
}

type errors struct {
	w                          *window
	vSplit, height, reportLine int
}

func (t *errors) Token(token parser.ParsedToken) {
	if token.Diagnostic != "" {
		message := fmt.Sprintf("%d:%d %s", token.Line+1, token.StartColumn+1, token.Diagnostic)
		emitStr(t.w.screen, 1, t.reportLine+2, mainStyle, message)
		t.reportLine++
	}
}

func (t *errors) Done() {

}

func (w *window) showPalette() {
	_, height := w.screen.Size()
	mainStyle := tcell.StyleDefault.Background(colorDeepBlue).Foreground(tcell.ColorWhite)
	i := 2
	j := 2
	sl := make([]tcell.Color, 0, len(tcell.ColorValues))
	for c := range tcell.ColorValues {
		sl = append(sl, c)
	}
	sort.Slice(sl, func(i, j int) bool {
		return sl[i].Hex() < sl[j].Hex()
	})
	for _, c := range sl {
		emitStr(w.screen, j, i, mainStyle.Foreground(c), fmt.Sprintf("%06x", c.Hex()))
		i++
		if i > height-2 {
			i = 2
			j += 8
		}
	}
}

var defStyle tcell.Style

// var menuStyle tcell.Style = defStyle.Background(tcell.ColorAqua)
var menuStyle tcell.Style = defStyle.Background(tcell.ColorSilver)

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	stl := style
	for _, c := range str {
		s.SetContent(x, y, c, nil, stl)
		x++
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, r rune) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, style)
		s.SetContent(col, y2, tcell.RuneHLine, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, style)
		s.SetContent(x2, row, tcell.RuneVLine, nil, style)
	}
	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, style)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, style)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, style)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, style)
	}
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			s.SetContent(col, row, r, nil, style)
		}
	}
}

var colorDeepBlue = tcell.NewRGBColor(0, 0, 63)

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

func splitLines(text string) [][]rune {
	lines := strings.Split(text, "\n")
	result := make([][]rune, len(lines))
	for i, line := range lines {
		result[i] = []rune(line)
	}
	return result
}
