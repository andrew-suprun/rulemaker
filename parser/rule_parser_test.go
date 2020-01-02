package parser

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"league.com/rulemaker/tokenizer"
)

func TestParser(t *testing.T) {
	for i, params := range fixture {
		lsnr := &listener{t: t, fixture: i, rules: params.rules, expected: params.tokens, debug: false}
		Parse(splitLines(params.rules), nil, nil, lsnr)
	}
}

type listener struct {
	debug    bool
	t        *testing.T
	fixture  int
	rules    string
	expected []ParsedToken
	got      []ParsedToken
}

func (l *listener) Token(token ParsedToken) {
	l.got = append(l.got, token)
}

func (l *listener) Done() {
	if l.debug {
		fmt.Printf("	{%q, []ParsedToken{\n", l.rules)
		for _, got := range l.got {
			if got.Diagnostic == "" {
				fmt.Printf("        {Token: tokenizer.Token{tokenizer.%s, %d, %d, %d, %q, %s}},\n",
					got.Type, got.Line, got.StartColumn, got.EndColumn, got.Text, value(got.Value))
			} else {
				fmt.Printf("        {Token: tokenizer.Token{tokenizer.%s, %d, %d, %d, %q, %s}, Diagnostic: %q},\n",
					got.Type, got.Line, got.StartColumn, got.EndColumn, got.Text, value(got.Value), got.Diagnostic)
			}
		}
		fmt.Println("	}},")
		return
	}
	for i, expected := range l.expected {
		if i >= len(l.got) {
			log.Println("fixture  ", l.fixture)
			log.Println("rules    ", l.rules)
			log.Println("expected ", expected)
			log.Println("got      nothing")
			l.t.FailNow()
		}
		got := l.got[i]
		if !reflect.DeepEqual(got, expected) {
			log.Println("fixture ", l.fixture)
			log.Println("rules    ", l.rules)
			log.Println("expected", expected)
			log.Println("got     ", got)
			l.t.FailNow()
		}
	}
}

func value(v interface{}) string {
	if v == nil {
		return "nil"
	}
	if str, ok := v.(string); ok {
		return fmt.Sprintf("%q", str)
	}
	return fmt.Sprintf("%v", v)
}

func splitLines(text string) [][]rune {
	lines := strings.Split(text, "\n")
	result := make([][]rune, len(lines))
	for i, line := range lines {
		result[i] = []rune(line)
	}
	return result
}

var fixture = []struct {
	rules  string
	tokens []ParsedToken
}{
	{"", []ParsedToken{}},
	{"# comment", []ParsedToken{
		{Token: tokenizer.Token{tokenizer.Comment, 0, 0, 9, `# comment`, nil}},
	}},
	{"abc = 1;", []ParsedToken{
		{Token: tokenizer.Token{tokenizer.CanonicalField, 0, 0, 3, "abc", "abc"}},
		{Token: tokenizer.Token{tokenizer.EqualSign, 0, 4, 5, "=", nil}},
		{Token: tokenizer.Token{tokenizer.IntegerLiteral, 0, 6, 7, "1", 1}},
		{Token: tokenizer.Token{tokenizer.Semicolon, 0, 7, 8, ";", nil}},
	}},
	{"foo=1;foo=2;", []ParsedToken{
		{Token: tokenizer.Token{tokenizer.CanonicalField, 0, 0, 3, "foo", "foo"}},
		{Token: tokenizer.Token{tokenizer.EqualSign, 0, 3, 4, "=", nil}},
		{Token: tokenizer.Token{tokenizer.IntegerLiteral, 0, 4, 5, "1", 1}},
		{Token: tokenizer.Token{tokenizer.Semicolon, 0, 5, 6, ";", nil}},
		{Token: tokenizer.Token{tokenizer.CanonicalField, 0, 6, 9, "foo", "foo"}, Diagnostic: "Redefinition of \"foo\"; previously defined at 1:1"},
		{Token: tokenizer.Token{tokenizer.EqualSign, 0, 9, 10, "=", nil}},
		{Token: tokenizer.Token{tokenizer.IntegerLiteral, 0, 10, 11, "2", 2}},
		{Token: tokenizer.Token{tokenizer.Semicolon, 0, 11, 12, ";", nil}},
	}},
	// TODO: validate fields/inputs/functions
	// TODO: incomplete last rule
	// TODO: comments after last rule
	// TODO: multiple equal signs
	// TODO: multiple semicolons
}
