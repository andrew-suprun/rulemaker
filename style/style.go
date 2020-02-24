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
	return result.Background(tcell.Color17)
}

func darkStyle(tokenType tokenizer.TokenType) tcell.Style {
	result := mainStyles[tokenType]
	return result.Background(tcell.Color235)
}

func lightStyle(tokenType tokenizer.TokenType) tcell.Style {
	result := lightStyles[tokenType]
	return result.Background(tcell.Color231)
}

var defStyle tcell.Style
var blueTheme = defStyle.Background(tcell.NewRGBColor(0, 0, 63))

var mainStyles = map[tokenizer.TokenType]tcell.Style{
	tokenizer.CanonicalField:   defStyle.Foreground(tcell.Color231),
	tokenizer.Operation:        defStyle.Foreground(tcell.Color87).Bold(true),
	tokenizer.Variable:         defStyle.Foreground(tcell.Color231),
	tokenizer.Label:            defStyle.Foreground(tcell.ColorTurquoise),
	tokenizer.Input:            defStyle.Foreground(tcell.ColorGreenYellow),
	tokenizer.OpenParenthesis:  defStyle.Foreground(tcell.Color231).Bold(true),
	tokenizer.CloseParenthesis: defStyle.Foreground(tcell.Color231).Bold(true),
	tokenizer.EqualSign:        defStyle.Foreground(tcell.Color231).Bold(true),
	tokenizer.Semicolon:        defStyle.Foreground(tcell.Color231).Bold(true),
	tokenizer.Comment:          defStyle.Foreground(tcell.Color248),
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
	tokenizer.CanonicalField:   defStyle.Foreground(tcell.ColorBlack),
	tokenizer.Operation:        defStyle.Foreground(tcell.Color18).Bold(true),
	tokenizer.Variable:         defStyle.Foreground(tcell.ColorBlack),
	tokenizer.Label:            defStyle.Foreground(tcell.Color21),
	tokenizer.Input:            defStyle.Foreground(tcell.ColorDarkGreen),
	tokenizer.OpenParenthesis:  defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.CloseParenthesis: defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.EqualSign:        defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.Semicolon:        defStyle.Foreground(tcell.ColorBlack).Bold(true),
	tokenizer.Comment:          defStyle.Foreground(tcell.Color245),
	tokenizer.StringLiteral:    defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.IntegerLiteral:   defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.RealLiteral:      defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.BooleanLiteral:   defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.NilLiteral:       defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.DateLiteral:      defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.YearSpanLiteral:  defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.MonthSpanLiteral: defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.DaySpanLiteral:   defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.TodayLiteral:     defStyle.Foreground(tcell.ColorRebeccaPurple),
	tokenizer.InvalidToken:     defStyle.Foreground(tcell.ColorRed).Bold(true),
}
