package tokenizer

import (
	"log"
	"reflect"
	"testing"
	"time"
)

func TestLineTokenizer(t *testing.T) {
	for _, f := range fixture {
		got := TokenizeString(f.line)
		got = got[:len(got)-1]
		if !reflect.DeepEqual(got, f.expected) {
			log.Println("line    ", f.line)
			log.Println("got     ", got)
			log.Println("expected", f.expected)
			t.FailNow()
		}
	}
}

var fixture = []struct {
	line     string
	expected Tokens
}{
	{"#c1\na#c2\n=#c3\n123;", Tokens{
		{0, 0, 3, Comment, nil},
		{1, 0, 1, CanonicalField, "a"},
		{1, 1, 4, Comment, nil},
		{2, 0, 1, EqualSign, nil},
		{2, 1, 4, Comment, nil},
		{3, 0, 3, IntegerLiteral, 123},
		{3, 3, 4, Semicolon, nil},
	}},

	{`(foo bar)`, Tokens{
		{0, 0, 1, OpenParenthesis, nil},
		{0, 1, 4, Operation, `foo`},
		{0, 5, 8, CanonicalField, `bar`},
		{0, 8, 9, CloseParenthesis, nil},
	}},
	{`"abc"`, Tokens{{0, 0, 5, StringLiteral, `abc`}}},
	{`"abc`, Tokens{{0, 0, 4, InvalidToken, nil}}},
	{`  "\""`, Tokens{{0, 2, 6, StringLiteral, `"`}}},
	{`"\`, Tokens{{0, 0, 2, InvalidToken, nil}}},
	{`""`, Tokens{{0, 0, 2, StringLiteral, ""}}},
	{`   # abc   `, Tokens{{0, 3, 11, Comment, nil}}},
	{` (())`, Tokens{
		{0, 1, 2, OpenParenthesis, nil},
		{0, 2, 3, OpenParenthesis, nil},
		{0, 3, 4, CloseParenthesis, nil},
		{0, 4, 5, CloseParenthesis, nil},
	}},
	{` @2020-01-02 `, Tokens{{0, 1, 12, DateLiteral, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-02`, Tokens{{0, 0, 11, DateLiteral, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-0 `, Tokens{{0, 0, 10, InvalidToken, nil}}},
	{`_`, Tokens{{0, 0, 1, Variable, `_`}}},
	{`  _abc_123_ `, Tokens{{0, 2, 11, Variable, `_abc_123_`}}},
	{`abc`, Tokens{{0, 0, 3, CanonicalField, `abc`}}},
	{` abc.+.-.123 `, Tokens{{0, 1, 12, CanonicalField, `abc.+.-.123`}}},
	{` today `, Tokens{{0, 1, 6, TodayLiteral, nil}}},
	{`nil`, Tokens{{0, 0, 3, NilLiteral, nil}}},
	{`true`, Tokens{{0, 0, 4, BooleanLiteral, true}}},
	{`false`, Tokens{{0, 0, 5, BooleanLiteral, false}}},
	{`  -00123  `, Tokens{{0, 2, 8, IntegerLiteral, -123}}},
	{`  -00123.  `, Tokens{{0, 2, 9, RealLiteral, -123.0}}},
	{`1y 2m 3d`, Tokens{
		{0, 0, 2, YearSpanLiteral, 1},
		{0, 3, 5, MonthSpanLiteral, 2},
		{0, 6, 8, DaySpanLiteral, 3},
	}},
	{`123 abc`, Tokens{{0, 0, 3, IntegerLiteral, 123}, {0, 4, 7, CanonicalField, `abc`}}},
	{`a = b;`, Tokens{
		{0, 0, 1, CanonicalField, `a`},
		{0, 2, 3, EqualSign, nil},
		{0, 4, 5, CanonicalField, `b`},
		{0, 5, 6, Semicolon, nil},
	}},
	{`a = (foo bar);`, Tokens{
		{0, 0, 1, CanonicalField, `a`},
		{0, 2, 3, EqualSign, nil},
		{0, 4, 5, OpenParenthesis, nil},
		{0, 5, 8, Operation, `foo`},
		{0, 9, 12, CanonicalField, `bar`},
		{0, 12, 13, CloseParenthesis, nil},
		{0, 13, 14, Semicolon, nil},
	}},
	{`a = (= bar);`, Tokens{
		{0, 0, 1, CanonicalField, `a`},
		{0, 2, 3, EqualSign, nil},
		{0, 4, 5, OpenParenthesis, nil},
		{0, 5, 6, Operation, `=`},
		{0, 7, 10, CanonicalField, `bar`},
		{0, 10, 11, CloseParenthesis, nil},
		{0, 11, 12, Semicolon, nil},
	}},
}
