package tokenizer

import (
	"log"
	"reflect"
	"testing"
	"time"

	"league.com/rulemaker/model"
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
		{model.Text{"#c1", 0, 0}, Comment, nil},
		{model.Text{"a", 1, 0}, CanonicalField, "a"},
		{model.Text{"#c2", 1, 1}, Comment, nil},
		{model.Text{"=", 2, 0}, EqualSign, nil},
		{model.Text{"#c3", 2, 1}, Comment, nil},
		{model.Text{"123", 3, 0}, IntegerLiteral, 123},
		{model.Text{";", 3, 3}, Semicolon, nil},
	}},

	{`(foo bar)`, Tokens{
		{model.Text{`(`, 0, 0}, OpenParen, nil},
		{model.Text{`foo`, 0, 1}, Operation, `foo`},
		{model.Text{`bar`, 0, 5}, CanonicalField, `bar`},
		{model.Text{`)`, 0, 8}, CloseParen, nil},
	}},
	{`"abc"`, Tokens{{model.Text{`"abc"`, 0, 0}, StringLiteral, `abc`}}},
	{`"abc`, Tokens{{model.Text{`"abc`, 0, 0}, InvalidToken, nil}}},
	{`  "\""`, Tokens{{model.Text{`"\""`, 0, 2}, StringLiteral, `"`}}},
	{`"\`, Tokens{{model.Text{`"\`, 0, 0}, InvalidToken, nil}}},
	{`""`, Tokens{{model.Text{`""`, 0, 0}, StringLiteral, ""}}},
	{`   # abc   `, Tokens{{model.Text{`# abc   `, 0, 3}, Comment, nil}}},
	{` (())`, Tokens{
		{model.Text{`(`, 0, 1}, OpenParen, nil},
		{model.Text{`(`, 0, 2}, OpenParen, nil},
		{model.Text{`)`, 0, 3}, CloseParen, nil},
		{model.Text{`)`, 0, 4}, CloseParen, nil},
	}},
	{` @2020-01-02 `, Tokens{{model.Text{`@2020-01-02`, 0, 1}, DateLiteral, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-02`, Tokens{{model.Text{`@2020-01-02`, 0, 0}, DateLiteral, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-0 `, Tokens{{model.Text{`@2020-01-0`, 0, 0}, InvalidToken, nil}}},
	{`_`, Tokens{{model.Text{`_`, 0, 0}, Variable, `_`}}},
	{`  _abc_123_ `, Tokens{{model.Text{`_abc_123_`, 0, 2}, Variable, `_abc_123_`}}},
	{`abc`, Tokens{{model.Text{`abc`, 0, 0}, CanonicalField, `abc`}}},
	{` abc.+.-.123 `, Tokens{{model.Text{`abc.+.-.123`, 0, 1}, CanonicalField, `abc.+.-.123`}}},
	{` today `, Tokens{{model.Text{`today`, 0, 1}, TodayLiteral, nil}}},
	{`nil`, Tokens{{model.Text{`nil`, 0, 0}, NilLiteral, nil}}},
	{`true`, Tokens{{model.Text{`true`, 0, 0}, BooleanLiteral, true}}},
	{`false`, Tokens{{model.Text{`false`, 0, 0}, BooleanLiteral, false}}},
	{`  -00123  `, Tokens{{model.Text{`-00123`, 0, 2}, IntegerLiteral, -123}}},
	{`  -00123.  `, Tokens{{model.Text{`-00123.`, 0, 2}, RealLiteral, -123.0}}},
	{`1y 2m 3d`, Tokens{
		{model.Text{`1y`, 0, 0}, YearSpanLiteral, 1},
		{model.Text{`2m`, 0, 3}, MonthSpanLiteral, 2},
		{model.Text{`3d`, 0, 6}, DaySpanLiteral, 3},
	}},
	{`123 abc`, Tokens{{model.Text{`123`, 0, 0}, IntegerLiteral, 123}, {model.Text{`abc`, 0, 4}, CanonicalField, `abc`}}},
	{`a = b;`, Tokens{
		{model.Text{`a`, 0, 0}, CanonicalField, `a`},
		{model.Text{`=`, 0, 2}, EqualSign, nil},
		{model.Text{`b`, 0, 4}, CanonicalField, `b`},
		{model.Text{`;`, 0, 5}, Semicolon, nil},
	}},
	{`a = (foo bar);`, Tokens{
		{model.Text{`a`, 0, 0}, CanonicalField, `a`},
		{model.Text{`=`, 0, 2}, EqualSign, nil},
		{model.Text{`(`, 0, 4}, OpenParen, nil},
		{model.Text{`foo`, 0, 5}, Operation, `foo`},
		{model.Text{`bar`, 0, 9}, CanonicalField, `bar`},
		{model.Text{`)`, 0, 12}, CloseParen, nil},
		{model.Text{`;`, 0, 13}, Semicolon, nil},
	}},
	{`a = (= bar);`, Tokens{
		{model.Text{`a`, 0, 0}, CanonicalField, `a`},
		{model.Text{`=`, 0, 2}, EqualSign, nil},
		{model.Text{`(`, 0, 4}, OpenParen, nil},
		{model.Text{`=`, 0, 5}, Operation, `=`},
		{model.Text{`bar`, 0, 7}, CanonicalField, `bar`},
		{model.Text{`)`, 0, 10}, CloseParen, nil},
		{model.Text{`;`, 0, 11}, Semicolon, nil},
	}},
}
