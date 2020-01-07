package tokenizer

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
	"unicode"
)

func Tokenize(content [][]rune) Tokens {
	tokens := make(Tokens, len(content))
	for i, line := range content {
		tokens[i] = TokenizeLine(line)
	}
	for {
		lastLine := tokens[len(tokens)-1]
		if len(lastLine) == 0 {
			tokens = tokens[:len(tokens)-1]
		} else {
			break
		}
	}
	return tokens
}

func TokenizeLine(runes []rune) TokenLine {
	tokenizer := &tokenizer{
		runes: runes,
	}
	return tokenizer.tokenizeLine()
}

type Tokens []TokenLine

type TokenLine []Token

type Token struct {
	Type   TokenType
	Column int
	Text   string
	Value  interface{}
}

func (t Token) String() string {
	return fmt.Sprintf("<%s %q column:%d value:%v>", t.Type, t.Text, t.Column, t.Value)
}

type TokenType int

const (
	InvalidToken TokenType = iota
	Identifier
	Variable
	Input
	Label
	StringLiteral
	IntegerLiteral
	RealLiteral
	BooleanLiteral
	NilLiteral
	DateLiteral
	DaySpanLiteral
	MonthSpanLiteral
	YearSpanLiteral
	TodayLiteral
	EqualSign
	Semicolon
	OpenParen
	CloseParen
	Comment
)

func (t TokenType) String() string {
	switch t {
	case InvalidToken:
		return "InvalidToken"
	case Identifier:
		return "Identifier"
	case Variable:
		return "Variable"
	case Input:
		return "Input"
	case Label:
		return "Label"
	case StringLiteral:
		return "StringLiteral"
	case IntegerLiteral:
		return "IntegerLiteral"
	case RealLiteral:
		return "RealLiteral"
	case BooleanLiteral:
		return "BooleanLiteral"
	case NilLiteral:
		return "NilLiteral"
	case DateLiteral:
		return "DateLiteral"
	case DaySpanLiteral:
		return "DaySpanLiteral"
	case MonthSpanLiteral:
		return "MonthSpanLiteral"
	case YearSpanLiteral:
		return "YearSpanLiteral"
	case TodayLiteral:
		return "TodayLiteral"
	case EqualSign:
		return "EqualSign"
	case Semicolon:
		return "Semicolon"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case Comment:
		return "Comment"
	}
	return "UnknownType"
}

type tokenizer struct {
	runes  []rune
	column int
	tokens TokenLine
}

func (t *tokenizer) tokenizeLine() TokenLine {
	t.column = 0
	for {
		t.skipSpace()
		if t.column >= len(t.runes) {
			return t.tokens
		}
		ch := t.runes[t.column]
		switch ch {
		case '#':
			t.comment()
		case '"':
			t.stringLiteral()
		case '=':
			t.equalSign()
		case ';':
			t.semicolon()
		case '(':
			t.openParen()
		case ')':
			t.closeParen()
		default:
			t.regularToken()
		}
	}
}

func (t *tokenizer) regularToken() {
	startColumn := t.column
	t.skipToSeparator()
	tokenText := t.runes[startColumn:t.column]
	token := string(tokenText)
	firstRune := tokenText[0]
	lastRune := tokenText[len(tokenText)-1]
	if lastRune == ':' {
		t.token(Label, startColumn, token)
		return
	}
	if lastRune == 'y' || lastRune == 'm' || lastRune == 'd' {
		intValue, err := strconv.ParseInt(string(tokenText[:len(tokenText)-1]), 10, 64)
		if err == nil {
			switch lastRune {
			case 'y':
				t.token(YearSpanLiteral, startColumn, int(intValue))
			case 'm':
				t.token(MonthSpanLiteral, startColumn, int(intValue))
			case 'd':
				t.token(DaySpanLiteral, startColumn, int(intValue))
			}
			return
		}
	}
	intValue, err := strconv.ParseInt(string(tokenText), 10, 64)
	if err == nil {
		t.token(IntegerLiteral, startColumn, int(intValue))
		return
	}

	floatValue, err := strconv.ParseFloat(string(tokenText), 64)
	if err == nil {
		t.token(RealLiteral, startColumn, floatValue)
		return
	}

	switch firstRune {
	case '@':
		date, err := time.Parse("2006-01-02", string(tokenText[1:]))
		if err == nil {
			t.token(DateLiteral, startColumn, date)
		} else {
			t.token(InvalidToken, startColumn, nil)
		}
		return
	case '_':
		t.token(Variable, startColumn, string(tokenText))
		return
	case '$':
		t.token(Input, startColumn, string(tokenText[1:]))
		return
	}
	switch token {
	case "true":
		t.token(BooleanLiteral, startColumn, true)
	case "false":
		t.token(BooleanLiteral, startColumn, false)
	case "nil":
		t.token(NilLiteral, startColumn, nil)
	case "today":
		t.token(TodayLiteral, startColumn, nil)
	default:
		t.token(Identifier, startColumn, token)
	}
}

func (t *tokenizer) comment() {
	startColumn := t.column
	t.column = len(t.runes)
	t.token(Comment, startColumn, nil)
}

func (t *tokenizer) stringLiteral() {
	startColumn := t.column
	t.column++
	escape := false
	closed := false
	buf := bytes.Buffer{}
loop:
	for ; t.column < len(t.runes); t.column++ {
		ch := t.runes[t.column]
		switch ch {
		case '\n':
			escape = true
			break loop
		case '\\':
			escape = true
		case '"':
			if !escape {
				closed = true
				t.column++
				break loop
			}
			buf.WriteRune(ch)
			escape = false
		default:
			buf.WriteRune(ch)
			escape = false
		}
	}

	if escape || !closed {
		t.token(InvalidToken, startColumn, nil)
		return
	}

	t.token(StringLiteral, startColumn, buf.String())
}

func (t *tokenizer) equalSign() {
	startColumn := t.column
	t.column++
	t.token(EqualSign, startColumn, nil)
}

func (t *tokenizer) semicolon() {
	startColumn := t.column
	t.column++
	t.token(Semicolon, startColumn, nil)
}

func (t *tokenizer) openParen() {
	startColumn := t.column
	t.column++
	t.token(OpenParen, startColumn, nil)
}

func (t *tokenizer) closeParen() {
	startColumn := t.column
	t.column++
	t.token(CloseParen, startColumn, nil)
}

func (t *tokenizer) token(tokenType TokenType, startColumn int, value interface{}) {
	t.tokens = append(t.tokens, Token{
		Type:   tokenType,
		Column: startColumn,
		Text:   string(t.runes[startColumn:t.column]),
		Value:  value,
	})
}

func (t *tokenizer) skipSpace() {
	for ; t.column < len(t.runes) && unicode.IsSpace(t.runes[t.column]); t.column++ {
	}
}

func (t *tokenizer) skipToSeparator() {
	for ; t.column < len(t.runes); t.column++ {
		ch := t.runes[t.column]
		if ch == '=' || ch == '(' || ch == ')' || ch == ';' || unicode.IsSpace(ch) {
			return
		}
	}
}
