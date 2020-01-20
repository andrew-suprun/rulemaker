package tokenizer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func TokenizeString(text string) Tokens {
	lines := strings.Split(text, "\n")
	runes := make([][]rune, len(lines))
	for i, line := range lines {
		runes[i] = []rune(line)
	}
	return TokenizeRunes(runes)
}

func TokenizeRunes(content [][]rune) Tokens {
	t := &tokenizer{}
	for line, runes := range content {
		t.tokenizeLine(line, runes)
	}
	t.tokens = append(t.tokens, Token{Type: EndMarker})
	return t.tokens
}

type Tokens []Token

type Token struct {
	Text   string
	Line   int
	Column int
	Type   TokenType
	Value  interface{}
}

func (t Token) String() string {
	return fmt.Sprintf("<%s %q %d:%d value=%v>", t.Type, t.Text, t.Line, t.Column, t.Value)
}

func (t Token) After(line, column int) bool {
	return t.Line > line || (t.Line == line && t.Column > column)
}

func (t Token) Contains(line, column int) bool {
	if t.Type == Comment {
		return line == t.Line && column >= t.Column
	}
	return line == t.Line && (column >= t.Column && column < t.Column+len(t.Text))
}

type TokenType int

const (
	InvalidToken TokenType = iota
	CanonicalField
	Operation
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
	OpenParenthesis
	CloseParenthesis
	Comment
	EndMarker
)

func (t TokenType) String() string {
	switch t {
	case InvalidToken:
		return "InvalidToken"
	case CanonicalField:
		return "CanonicalField"
	case Operation:
		return "Operation"
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
	case OpenParenthesis:
		return "OpenParenthesis"
	case CloseParenthesis:
		return "CloseParenthesis"
	case Comment:
		return "Comment"
	case EndMarker:
		return "EndMarker"
	}
	return "UnknownType"
}

type tokenizer struct {
	runes         []rune
	line          int
	column        int
	tokens        Tokens
	lastTokenType TokenType
}

func (t *tokenizer) tokenizeLine(line int, runes []rune) {
	t.runes = runes
	t.line = line
	t.column = 0
	for {
		t.skipSpace()
		if t.column >= len(t.runes) {
			return
		}
		ch := t.runes[t.column]
		switch ch {
		case '#':
			t.comment()
		case '"':
			t.stringLiteral()
		case '=':
			if t.lastTokenType != OpenParenthesis {
				t.equalSign()
			} else {
				t.regularToken()
			}
		case ';':
			t.semicolon()
		case '(':
			t.openParenthesis()
		case ')':
			t.closeParenthesis()
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
		if t.lastTokenType == OpenParenthesis {
			t.token(Operation, startColumn, token)
		} else {
			t.token(CanonicalField, startColumn, token)
		}
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

func (t *tokenizer) openParenthesis() {
	startColumn := t.column
	t.column++
	t.token(OpenParenthesis, startColumn, nil)
}

func (t *tokenizer) closeParenthesis() {
	startColumn := t.column
	t.column++
	t.token(CloseParenthesis, startColumn, nil)
}

func (t *tokenizer) token(tokenType TokenType, startColumn int, value interface{}) {
	t.tokens = append(t.tokens, Token{
		Text:   string(t.runes[startColumn:t.column]),
		Line:   t.line,
		Column: startColumn,
		Type:   tokenType,
		Value:  value,
	})
	if tokenType != Comment {
		t.lastTokenType = tokenType
	}
}

func (t *tokenizer) skipSpace() {
	for ; t.column < len(t.runes) && unicode.IsSpace(t.runes[t.column]); t.column++ {
	}
}

func (t *tokenizer) skipToSeparator() {
	for ; t.column < len(t.runes); t.column++ {
		ch := t.runes[t.column]
		if ch == '(' || ch == ')' || ch == '"' || ch == ';' || ch == '#' || unicode.IsSpace(ch) {
			return
		}
	}
}
