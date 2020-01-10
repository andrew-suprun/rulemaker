package tokenizer

import (
	"log"
	"reflect"
	"testing"
	"time"
)

func TestLineTokenizer(t *testing.T) {
	for _, f := range fixture {
		tokenizer := &tokenizer{}
		tokenizer.tokenizeLine(0, []rune(f.line))
		if !reflect.DeepEqual(tokenizer.tokens, f.tokens) {
			log.Println("line    ", f.line)
			log.Println("got     ", tokenizer.tokens)
			log.Println("expected", f.tokens)
			t.FailNow()
		}
	}
}

var fixture = []struct {
	line   string
	tokens Tokens
}{
	{`(foo bar)`, Tokens{{OpenParen, 0, 0, `(`, nil}, {Operation, 0, 1, `foo`, `foo`}, {CanonicalField, 0, 5, `bar`, `bar`}, {CloseParen, 0, 8, `)`, nil}}},
	{`"abc"`, Tokens{{StringLiteral, 0, 0, `"abc"`, `abc`}}},
	{`"abc`, Tokens{{InvalidToken, 0, 0, `"abc`, nil}}},
	{`  "\""`, Tokens{{StringLiteral, 0, 2, `"\""`, `"`}}},
	{`"\`, Tokens{{InvalidToken, 0, 0, `"\`, nil}}},
	{`""`, Tokens{{StringLiteral, 0, 0, `""`, ""}}},
	{`   # abc   `, Tokens{{Comment, 0, 3, `# abc   `, nil}}},
	{` (())`, Tokens{{OpenParen, 0, 1, `(`, nil}, {OpenParen, 0, 2, `(`, nil}, {CloseParen, 0, 3, `)`, nil}, {CloseParen, 0, 4, `)`, nil}}},
	{` @2020-01-02 `, Tokens{{DateLiteral, 0, 1, `@2020-01-02`, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-02`, Tokens{{DateLiteral, 0, 0, `@2020-01-02`, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-0 `, Tokens{{InvalidToken, 0, 0, `@2020-01-0`, nil}}},
	{`_`, Tokens{{Variable, 0, 0, `_`, `_`}}},
	{`  _abc_123_ `, Tokens{{Variable, 0, 2, `_abc_123_`, `_abc_123_`}}},
	{`abc`, Tokens{{CanonicalField, 0, 0, `abc`, `abc`}}},
	{` abc.+.-.123 `, Tokens{{CanonicalField, 0, 1, `abc.+.-.123`, `abc.+.-.123`}}},
	{` today `, Tokens{{TodayLiteral, 0, 1, `today`, nil}}},
	{`nil`, Tokens{{NilLiteral, 0, 0, `nil`, nil}}},
	{`true`, Tokens{{BooleanLiteral, 0, 0, `true`, true}}},
	{`false`, Tokens{{BooleanLiteral, 0, 0, `false`, false}}},
	{`  -00123  `, Tokens{{IntegerLiteral, 0, 2, `-00123`, -123}}},
	{`  -00123.  `, Tokens{{RealLiteral, 0, 2, `-00123.`, -123.0}}},
	{`1y 2m 3d`, Tokens{{YearSpanLiteral, 0, 0, `1y`, 1}, {MonthSpanLiteral, 0, 3, `2m`, 2}, {DaySpanLiteral, 0, 6, `3d`, 3}}},
	{`123 abc`, Tokens{{IntegerLiteral, 0, 0, `123`, 123}, {CanonicalField, 0, 4, `abc`, `abc`}}},
	{`a = b;`, Tokens{{CanonicalField, 0, 0, `a`, `a`}, {EqualSign, 0, 2, `=`, nil}, {CanonicalField, 0, 4, `b`, `b`}, {Semicolon, 0, 5, `;`, nil}}},
	{`a = (foo bar);`, Tokens{{CanonicalField, 0, 0, `a`, `a`}, {EqualSign, 0, 2, `=`, nil}, {OpenParen, 0, 4, `(`, nil}, {Operation, 0, 5, `foo`, `foo`},
		{CanonicalField, 0, 9, `bar`, `bar`}, {CloseParen, 0, 12, `)`, nil}, {Semicolon, 0, 13, `;`, nil}}},
	{`a = (= bar);`, Tokens{{CanonicalField, 0, 0, `a`, `a`}, {EqualSign, 0, 2, `=`, nil}, {OpenParen, 0, 4, `(`, nil}, {Operation, 0, 5, `=`, `=`},
		{CanonicalField, 0, 7, `bar`, `bar`}, {CloseParen, 0, 10, `)`, nil}, {Semicolon, 0, 11, `;`, nil}}},
}
