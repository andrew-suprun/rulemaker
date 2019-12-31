package tokenizer

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func NewTokenizer(content string) Tokenizer {
	return newTokenizer(content)
}

type Tokenizer interface {
	Tokens() TokenLines
}

type TokenLines []TokenLine
type TokenLine []Token
type Token struct {
	Type        TokenType
	StartColumn int
	EndColumn   int
	Text        string
	Value       interface{}
}

func (t Token) String() string {
	return fmt.Sprintf("<token %d-%d: %q value:%v type:%s>", t.StartColumn, t.EndColumn, t.Text, t.Value, t.Type)
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
	FloatingPointLiteral
	BooleanLiteral
	NilLiteral
	DateLiteral
	DaySpanLiteral
	MonthSpanLiteral
	YearSpanLiteral
	TodayLiteral
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
	case FloatingPointLiteral:
		return "FloatingPointLiteral"
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
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case Comment:
		return "Comment"
	}
	return "UnknownType"
}

func newTokenizer(content string) Tokenizer {
	tokenizer := &tokenizer{
		content: splitLines(content),
	}
	tokenizer.tokenize()
	return tokenizer
}

func splitLines(text string) [][]rune {
	lines := strings.Split(text, "\n")
	result := make([][]rune, len(lines))
	for i, line := range lines {
		result[i] = []rune(line)
	}
	return result
}

func (t *tokenizer) tokenize() {
	t.tokenLines = make(TokenLines, len(t.content))
	for i, line := range t.content {
		t.tokenLines[i] = t.tokenizeLine(line)
	}
}

func (t *tokenizer) tokenizeLine(line []rune) TokenLine {
	lt := &lineTokenizer{line: line}
	return lt.tokenize()
}

type lineTokenizer struct {
	line   []rune
	column int
}

func (t *lineTokenizer) tokenize() (tokens TokenLine) {
	for {
		t.skipSpace()
		if t.column >= len(t.line) {
			return tokens
		}
		ch := t.line[t.column]
		switch ch {
		case '#':
			tokens = append(tokens, t.comment())
		case '"':
			tokens = append(tokens, t.stringLiteral())
		case '(':
			tokens = append(tokens, t.openParen())
		case ')':
			tokens = append(tokens, t.closeParen())
		default:
			startColumn := t.column
			t.skipToSeparator()
			tokenText := t.line[startColumn:t.column]
			tokenType, tokenValue := tokenTypeAndValue(tokenText)
			token := Token{
				Type:        tokenType,
				StartColumn: startColumn,
				EndColumn:   t.column,
				Text:        string(tokenText),
				Value:       tokenValue,
			}
			tokens = append(tokens, token)
		}
	}
}

func tokenTypeAndValue(tokenText []rune) (TokenType, interface{}) {
	token := string(tokenText)
	firstRune := tokenText[0]
	lastRune := tokenText[len(tokenText)-1]
	if lastRune == ':' {
		return Label, token
	}
	if lastRune == 'y' || lastRune == 'm' || lastRune == 'd' {
		intValue, err := strconv.ParseInt(string(tokenText[:len(tokenText)-1]), 10, 64)
		if err == nil {
			switch lastRune {
			case 'y':
				return YearSpanLiteral, int(intValue)
			case 'm':
				return MonthSpanLiteral, int(intValue)
			case 'd':
				return DaySpanLiteral, int(intValue)
			}
		}
	}
	intValue, err := strconv.ParseInt(string(tokenText), 10, 64)
	if err == nil {
		return IntegerLiteral, int(intValue)
	}

	floatValue, err := strconv.ParseFloat(string(tokenText), 64)
	if err == nil {
		return FloatingPointLiteral, floatValue
	}

	switch firstRune {
	case '@':
		date, err := time.Parse("2006-01-02", string(tokenText[1:]))
		if err != nil {
			return InvalidToken, nil
		}
		return DateLiteral, date
	case '_':
		return Variable, string(tokenText)
	case '$':
		return Input, string(tokenText[1:])
	}
	switch token {
	case "true":
		return BooleanLiteral, true
	case "false":
		return BooleanLiteral, false
	case "nil":
		return NilLiteral, nil
	case "today":
		return TodayLiteral, nil
	}

	return Identifier, token
}

func (t *tokenizer) Tokens() TokenLines {
	return t.tokenLines
}

type tokenizer struct {
	content    [][]rune
	tokenLines TokenLines
}

func (t *lineTokenizer) comment() Token {
	startColumn := t.column
	t.column = len(t.line)
	return t.token(Comment, startColumn, nil)
}

func (t *lineTokenizer) stringLiteral() Token {
	startColumn := t.column
	t.column++
	escape := false
	closed := false
	buf := bytes.Buffer{}
loop:
	for ; t.column < len(t.line); t.column++ {
		ch := t.line[t.column]
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
		return t.token(InvalidToken, startColumn, nil)
	}

	return t.token(StringLiteral, startColumn, buf.String())
}

func (t *lineTokenizer) openParen() Token {
	startColumn := t.column
	t.column++
	return t.token(OpenParen, startColumn, nil)
}

func (t *lineTokenizer) closeParen() Token {
	startColumn := t.column
	t.column++
	return t.token(CloseParen, startColumn, nil)
}

func (t *lineTokenizer) token(tokenType TokenType, startColumn int, value interface{}) (token Token) {
	return Token{
		Type:        tokenType,
		StartColumn: startColumn,
		EndColumn:   t.column,
		Text:        string(t.line[startColumn:t.column]),
		Value:       value,
	}
}

func (t *lineTokenizer) skipSpace() {
	for ; t.column < len(t.line) && unicode.IsSpace(t.line[t.column]); t.column++ {
	}
}

func (t *lineTokenizer) skipToSeparator() {
	for ; t.column < len(t.line); t.column++ {
		ch := t.line[t.column]
		if ch == '(' || ch == ')' || unicode.IsSpace(ch) {
			return
		}
	}
}
