package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"league.com/rulemaker/canonical_model"
	"league.com/rulemaker/meta"
	"league.com/rulemaker/parser"
	"league.com/rulemaker/tokenizer"

	"github.com/gdamore/tcell"
)

func main() {
	metainfo := meta.Metainfo(canonical_model.EmployeeDTO{})

	var inputs = parser.Set{
		"policy":                        {},
		"sin":                           {},
		"employee_id":                   {},
		"last_name":                     {},
		"given_names":                   {},
		"person_type":                   {},
		"effective_date":                {},
		"transaction_date":              {},
		"division":                      {},
		"benefit_class":                 {},
		"administrative_class":          {},
		"retirement_date":               {},
		"termination_date":              {},
		"deceased_date":                 {},
		"birth_date":                    {},
		"gender":                        {},
		"language":                      {},
		"street":                        {},
		"city":                          {},
		"province_state":                {},
		"postal_zip_code":               {},
		"foreign_country":               {},
		"hire_date":                     {},
		"province_of_employment":        {},
		"province_of_residence":         {},
		"employee_smoker":               {},
		"business_location":             {},
		"cost_centre":                   {},
		"tax_exempt":                    {},
		"does_employee_have_dependants": {},
		"spouse_or_common_law_spouse":   {},
		"num_of_dependants":             {},
		"bank_transit_id":               {},
		"bank_number":                   {},
		"bank_account_number":           {},
		"earnings_amount":               {},
		"earnings_frequency":            {},
		"dependant_name_on_drug_card":   {},
		"revision_reason":               {},
		"created_by":                    {},
	}

	var operations = parser.Set{
		"strip_prefix":        {},
		"strip_leading_zeros": {},
		"first_of":            {},
		"map":                 {},
		"select":              {},
		"all":                 {},
		"any":                 {},
		"one_of":              {},
		"join":                {},
		"+":                   {},
		"*":                   {},
		"=":                   {},
		"!=":                  {},
		"<":                   {},
		">":                   {},
		"<=":                  {},
		">=":                  {},
		"min":                 {},
		"max":                 {},
		"has":                 {},
		"first_of_month":      {},
		"weekly_hours":        {},
		"config":              {},
		"fail":                {},
		"log":                 {},
		"ticket":              {},
		"contains":            {},
		"skip":                {},
	}

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

	w := window{screen: s, content: splitLines(string(content)), metainfo: metainfo, inputs: inputs, operations: operations}

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
			}
		case *tcell.EventMouse:
			button := ev.Buttons()
			if button&tcell.WheelUp != 0 && lineOffset > 0 {
				lineOffset--
			}
			if button&tcell.WheelDown != 0 && lineOffset < len(w.content)-2 {
				lineOffset++
			}
		}
	}
}

type window struct {
	screen     tcell.Screen
	title      string
	status     string
	content    [][]rune
	metainfo   meta.Meta
	inputs     parser.Set
	operations parser.Set
}

var lineOffset = 0

var mainStyle = tcell.StyleDefault.Background(colorDeepBlue).Foreground(tcell.ColorWhite)
var lineNumberStyle = mainStyle.Foreground(tcell.ColorSilver).Background(tcell.ColorBlack)

func (w *window) draw() {
	width, height := w.screen.Size()
	var vSplit = 48 // TODO: make it width - 50

	for row := 2; row < height-1; row++ {
		for col := 5; col < width; col++ {
			w.screen.SetContent(col, row, ' ', nil, mainStyle)
		}
	}
	for row := 2; row < height-1; row++ {
		for col := 0; col < 5; col++ {
			w.screen.SetContent(col, row, ' ', nil, lineNumberStyle)
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

	tokenizer.Tokenize(w.content,
		parser.NewParser(w.metainfo, w.inputs, w.operations,
			newTee(
				// &errors{w: w, vSplit: vSplit, height: height}, // TODO: move right
				NewWrapper(5, vSplit,
					&tokens{w: w, vSplit: vSplit, height: height},
				),
			),
		),
	)

	w.screen.Show()
}

type PositionedTokens interface {
	Token(token PositionedToken)
	Done()
}

type PositionedToken struct {
	parser.ParsedToken
	X, Y int
}

func (t PositionedToken) String() string {
	return fmt.Sprintf("<positioned token: %s; position: %d:%d>", t.Token, t.Y, t.X)
}

type wrapper struct {
	lMargin, rMargin     int
	tokens               PositionedTokens
	lineOffset           int
	lastX, lastY         int
	lastColumn, lastLine int
}

func NewWrapper(lMargin, rMargin int, tokens PositionedTokens) parser.Tokens {
	return &wrapper{lMargin: lMargin, rMargin: rMargin, tokens: tokens}
}

func (w *wrapper) Token(token parser.ParsedToken) {
	outToken := PositionedToken{
		ParsedToken: token,
		X:           token.StartColumn - w.lastColumn + w.lastX,
		Y:           token.Line - w.lastLine + w.lastY, //+ w.lineOffset,
	}
	if outToken.X+len(token.Text) > w.rMargin {
		outToken.X = w.lMargin + 4
		outToken.Y++
		// w.lineOffset++
	}
	if w.lastLine != token.Line {
		outToken.X = w.lMargin + token.StartColumn
	}
	// log.Printf("### wrapper: outToken.X=%d  len(token.Text)=%d  w.rMargin=%d", outToken.X, len(token.Text), w.rMargin)
	w.lastX = outToken.X
	w.lastY = outToken.Y
	w.lastColumn = token.StartColumn
	w.lastLine = token.Line
	w.tokens.Token(outToken)
}

func (w *wrapper) Done() {
	w.tokens.Done()
}

type tokens struct {
	w              *window
	vSplit, height int
	shownLines     int
}

func (t *tokens) Token(token PositionedToken) {
	if token.Y < lineOffset || token.Y >= lineOffset+t.height-3 {
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
	for ; t.shownLines <= token.Y; t.shownLines++ {
		if t.shownLines < lineOffset {
			continue
		}
		emitStr(t.w.screen, 0, t.shownLines+2-lineOffset, lineNumberStyle, fmt.Sprintf("%4d ", t.shownLines+1))
	}

	emitStr(t.w.screen, token.X, token.Y+2-lineOffset, tokenStyle, token.Text)
}

func (t *tokens) Done() {
	// for line := lineOffset; line < lineOffset+t.height-3; line++ {
	// 	emitStr(t.w.screen, 0, line+2-lineOffset, style, fmt.Sprintf("%4d ", line+1))
	// }
}

type errors struct {
	w                          *window
	vSplit, height, reportLine int
}

func (t *errors) Token(token parser.ParsedToken) {
	if token.Diagnostic != "" {
		message := fmt.Sprintf("%d:%d %s", token.Line+1, token.StartColumn+1, token.Diagnostic)
		lines := wrapLines(message, t.vSplit)
		for _, line := range lines {
			emitStr(t.w.screen, 1, t.reportLine+2, mainStyle, line)
			t.reportLine++
		}
	}
}

func (t *errors) Done() {

}

func wrapLines(str string, w int) (result []string) {
	if len(str) > w {
		result = append(result, str[:w])
		str = str[w:]
	}
	for len(str) > w-4 {
		result = append(result, "    "+str[:w-4])
		str = str[w-4:]
	}
	result = append(result, "    "+str)
	return result
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
