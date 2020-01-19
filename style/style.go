package style

import (
	"github.com/gdamore/tcell"
	"league.com/rulemaker/tokenizer"
)

type Theme int

const (
	BlueTheme Theme = iota
	DarkTheme
	LightTheme
)

func TokenStyle(tokenType tokenizer.TokenType, theme Theme) tcell.Style {
	switch theme {
	case BlueTheme:
		return blueStyle(tokenType)
	case DarkTheme:
		return darkStyle(tokenType)
	case LightTheme:
		return lightStyle(tokenType)
	}
	return 0
}

func blueStyle(tokenType tokenizer.TokenType) tcell.Style {
	result := mainStyles[tokenType]
	return result.Background(tcell.NewRGBColor(0, 0, 63))
}

func darkStyle(tokenType tokenizer.TokenType) tcell.Style {
	result := mainStyles[tokenType]
	return result.Background(tcell.ColorBlack)
}

func lightStyle(tokenType tokenizer.TokenType) tcell.Style {
	return lightStyles[tokenType]
}

var defStyle tcell.Style
var blueTheme = defStyle.Background(tcell.NewRGBColor(0, 0, 63))

var mainStyles = map[tokenizer.TokenType]tcell.Style{
	tokenizer.CanonicalField:   defStyle.Foreground(tcell.ColorWhite).Bold(true),
	tokenizer.Operation:        defStyle.Foreground(tcell.ColorWhite).Bold(true),
	tokenizer.Variable:         defStyle.Foreground(tcell.ColorWhite),
	tokenizer.Label:            defStyle.Foreground(tcell.NewHexColor(0x8fffff)),
	tokenizer.Input:            defStyle.Foreground(tcell.NewHexColor(0x8fff8f)).Bold(true),
	tokenizer.OpenParen:        defStyle.Foreground(tcell.ColorWhite).Bold(true),
	tokenizer.CloseParen:       defStyle.Foreground(tcell.ColorWhite).Bold(true),
	tokenizer.EqualSign:        defStyle.Foreground(tcell.ColorWhite).Bold(true),
	tokenizer.Semicolon:        defStyle.Foreground(tcell.ColorWhite).Bold(true),
	tokenizer.Comment:          defStyle.Foreground(tcell.NewHexColor(0xadadad)),
	tokenizer.StringLiteral:    defStyle.Foreground(tcell.ColorGold),
	tokenizer.IntegerLiteral:   defStyle.Foreground(tcell.ColorGold),
	tokenizer.RealLiteral:      defStyle.Foreground(tcell.ColorGold),
	tokenizer.BooleanLiteral:   defStyle.Foreground(tcell.ColorGold),
	tokenizer.NilLiteral:       defStyle.Foreground(tcell.ColorGold),
	tokenizer.DateLiteral:      defStyle.Foreground(tcell.ColorGold),
	tokenizer.YearSpanLiteral:  defStyle.Foreground(tcell.ColorGold),
	tokenizer.MonthSpanLiteral: defStyle.Foreground(tcell.ColorGold),
	tokenizer.DaySpanLiteral:   defStyle.Foreground(tcell.ColorGold),
	tokenizer.TodayLiteral:     defStyle.Foreground(tcell.ColorGold),
	tokenizer.InvalidToken:     defStyle.Foreground(tcell.ColorRed).Bold(true),
}

var lightStyles = map[tokenizer.TokenType]tcell.Style{
	tokenizer.CanonicalField:   defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.Operation:        defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.Variable:         defStyle.Foreground(tcell.ColorBlack),
	tokenizer.Label:            defStyle.Foreground(tcell.NewHexColor(0x003f3f)),
	tokenizer.Input:            defStyle.Foreground(tcell.NewHexColor(0x003f00)).Bold(true),
	tokenizer.OpenParen:        defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.CloseParen:       defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.EqualSign:        defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.Semicolon:        defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.Comment:          defStyle.Foreground(tcell.NewHexColor(0x3d3d3d)),
	tokenizer.StringLiteral:    defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.IntegerLiteral:   defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.RealLiteral:      defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.BooleanLiteral:   defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.NilLiteral:       defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.DateLiteral:      defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.YearSpanLiteral:  defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.MonthSpanLiteral: defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.DaySpanLiteral:   defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.TodayLiteral:     defStyle.Foreground(tcell.NewHexColor(0x3f3f00)),
	tokenizer.InvalidToken:     defStyle.Foreground(tcell.NewHexColor(0x7f0000)).Bold(true).Bold(true),
}
