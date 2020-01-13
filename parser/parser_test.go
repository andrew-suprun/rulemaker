package parser

import (
	"log"
	"reflect"
	"strings"
	"testing"

	"league.com/rulemaker/meta"
	"league.com/rulemaker/model"
	"league.com/rulemaker/tokenizer"
)

func TestSplitRules(t *testing.T) {
	p := NewParser(
		meta.Meta{"foo": meta.Int, "bar": meta.String},
		model.Set{"x": {}, "y": {}},
		model.Set{"baz": {}, "quux": {}},
	).(*parser)

	for i, test := range starts {
		tokens := tokenizer.Tokenize(splitLines(test.line))
		p.Parse(tokens)
		if !reflect.DeepEqual(test.starts, p.ruleStarts) {
			log.Println("fixture  ", i)
			log.Println("line     ", test.line)
			log.Println("expected ", test.starts)
			log.Println("got      ", p.ruleStarts)
			t.FailNow()
		}
	}

}

type startsFixture struct {
	line   string
	starts []int
}

var starts = []startsFixture{
	{"# comment\n a #comment\n = #comment\n 123", []int{0, 1, 6}},
	{"=", []int{0, 1}},
	{"===", []int{0, 3}},
	{"a = = d", []int{0, 4}},
	{"a = b = d", []int{0, 2, 5}},
	{"a = true = d", []int{0, 5}},
	{"a = b c = d", []int{0, 3, 6}},
	{";=", []int{0, 1, 2}},
	{";;", []int{0, 1, 2}},
}

func TestParser(t *testing.T) {
	for i, params := range fixture {
		p := NewParser(
			meta.Meta{"foo": meta.Int, "bar": meta.String},
			model.Set{"x": {}, "y": {}},
			model.Set{"baz": {}, "quux": {}})
		p.Parse(tokenizer.Tokenize(splitLines(params.rules)))
		got := p.Diagnostics()

		for j, expected := range params.errors {
			if j >= len(got) {
				log.Println("fixture  ", i)
				log.Println("rules    ", params.rules)
				log.Println("expected ", expected)
				log.Println("got      nothing")
				t.FailNow()
			}
			if !reflect.DeepEqual(got[j], expected) {
				log.Println("fixture ", i)
				log.Println("rules    ", params.rules)
				log.Println("expected ", expected)
				log.Println("got      ", got[j])
				t.FailNow()
			}
		}
		if len(got) > len(params.errors) {
			log.Println("fixture  ", i)
			log.Println("rules    ", params.rules)
			log.Println("expected nothing")
			log.Println("got      ", got[len(params.errors)])
			t.FailNow()
		}
	}
}

var fixture = []struct {
	rules  string
	errors []Diagnostic
}{
	{"", nil},
	{"# comment", nil},
	{"abc = 1;", []Diagnostic{
		{0, 0, "Canonical model does not have field \"abc\""},
	}},
	{"foo = 1;# comment", nil},
	{"foo = 1", nil},
	{"foo = 1;foo = 2;", []Diagnostic{
		{0, 8, `Canonical field "foo" redefined; previously defined at 1:1`},
	}},
	{"foo = (baz 123);", nil},
	{"foo = (unknown 123);", []Diagnostic{
		{0, 7, "Operation \"unknown\" is not defined"},
	}},
	{"foo = = 1y;", []Diagnostic{
		{0, 6, "Unexpected '='"},
	}},
	{"foo = 1m;;;", nil},
	{"bar = foo;", []Diagnostic{
		{0, 6, "Canonical field \"foo\" is not defined"},
	}},
	{"foo = 1m;bar = foo;", nil},
	{"foo = $x; bar = $a;", []Diagnostic{
		{0, 16, "Input field \"$a\" is not defined"},
	}},
	{"foo = (((;", []Diagnostic{
		{0, 6, "Missing operation"},
		{0, 7, "Missing operation"},
		{0, 8, "Unbalanced '('"},
		{0, 9, "Missing operation"},
	}},
}

func splitLines(text string) [][]rune {
	lines := strings.Split(text, "\n")
	result := make([][]rune, len(lines))
	for i, line := range lines {
		result[i] = []rune(line)
	}
	return result
}
