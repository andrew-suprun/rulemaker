package tokenizer

import (
	"log"
	"reflect"
	"testing"
	"time"
)

func TestLineTokenizer(t *testing.T) {
	for _, f := range fixture {
		tokens := TokenizeLine([]rune(f.line))
		if !reflect.DeepEqual(tokens, f.tokens) {
			log.Println("line    ", f.line)
			log.Println("got     ", tokens)
			log.Println("expected", f.tokens)
			t.FailNow()
		}
	}
}

var fixture = []struct {
	line   string
	tokens TokenLine
}{
	{`(foo bar)`, TokenLine{{OpenParen, 0, `(`, nil}, {Identifier, 1, `foo`, `foo`}, {Identifier, 5, `bar`, `bar`}, {CloseParen, 8, `)`, nil}}},
	{`"abc"`, TokenLine{{StringLiteral, 0, `"abc"`, `abc`}}},
	{`"abc`, TokenLine{{InvalidToken, 0, `"abc`, nil}}},
	{`  "\""`, TokenLine{{StringLiteral, 2, `"\""`, `"`}}},
	{`"\`, TokenLine{{InvalidToken, 0, `"\`, nil}}},
	{`""`, TokenLine{{StringLiteral, 0, `""`, ""}}},
	{`   # abc   `, TokenLine{{Comment, 3, `# abc   `, nil}}},
	{` (())`, TokenLine{{OpenParen, 1, `(`, nil}, {OpenParen, 2, `(`, nil}, {CloseParen, 3, `)`, nil}, {CloseParen, 4, `)`, nil}}},
	{` @2020-01-02 `, TokenLine{{DateLiteral, 1, `@2020-01-02`, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-02`, TokenLine{{DateLiteral, 0, `@2020-01-02`, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-0 `, TokenLine{{InvalidToken, 0, `@2020-01-0`, nil}}},
	{`_`, TokenLine{{Variable, 0, `_`, `_`}}},
	{`  _abc_123_ `, TokenLine{{Variable, 2, `_abc_123_`, `_abc_123_`}}},
	{`abc`, TokenLine{{Identifier, 0, `abc`, `abc`}}},
	{` abc.+.-.123 `, TokenLine{{Identifier, 1, `abc.+.-.123`, `abc.+.-.123`}}},
	{` today `, TokenLine{{TodayLiteral, 1, `today`, nil}}},
	{`nil`, TokenLine{{NilLiteral, 0, `nil`, nil}}},
	{`true`, TokenLine{{BooleanLiteral, 0, `true`, true}}},
	{`false`, TokenLine{{BooleanLiteral, 0, `false`, false}}},
	{`  -00123  `, TokenLine{{IntegerLiteral, 2, `-00123`, -123}}},
	{`  -00123.  `, TokenLine{{RealLiteral, 2, `-00123.`, -123.0}}},
	{`1y 2m 3d`, TokenLine{{YearSpanLiteral, 0, `1y`, 1}, {MonthSpanLiteral, 3, `2m`, 2}, {DaySpanLiteral, 6, `3d`, 3}}},
	{`123 abc`, TokenLine{{IntegerLiteral, 0, `123`, 123}, {Identifier, 4, `abc`, `abc`}}},
	{`a = b;`, TokenLine{{Identifier, 0, `a`, `a`}, {EqualSign, 2, `=`, nil}, {Identifier, 4, `b`, `b`}, {Semicolon, 5, `;`, nil}}},
}
