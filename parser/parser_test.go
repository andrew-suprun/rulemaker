package parser

import (
	"fmt"
	"log"
	"reflect"
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
	)

	for i, test := range starts {
		p.tokens = tokenizer.TokenizeString(test.line)
		// for i, token := range tokens {
		// 	fmt.Printf("token %d: %s\n", i, token)
		// }
		p.makeRules()
		got := fmt.Sprint(p.rules)
		if !reflect.DeepEqual(got, test.expected) {
			log.Println("fixture  ", i)
			log.Printf("line     %q\n", test.line)
			log.Printf("expected %#v\n", test.expected)
			log.Printf("got      %#v\n", got)
			t.FailNow()
		}
	}

}

type startsFixture struct {
	line     string
	expected string
}

type result struct {
	head, body []point
}

type point struct {
	line, column int
}

var starts = []startsFixture{
	{"", "[]"},
	{"123", "[<rule: 0-1-1>]"},
	{"a", "[<rule: 0-1-1 field: 0>]"},
	{"nil", "[<rule: 0-1-1>]"},
	{";", "[<rule: 0-0-1>]"},
	{"_abc = ;_", "[<rule: 0-1-3 field: 0> <rule: 3-4-4 field: 3>]"},
	{"#c1\n a#c2\n= #c3\n 123;", "[<rule: 0-3-7 field: 1>]"},
	{"=", "[<rule: 0-0-1>]"},
	{"=;", "[<rule: 0-0-2>]"},
	{"a = = d", "[<rule: 0-1-4 field: 0>]"},
	{"a = b = d", "[<rule: 0-1-5 field: 0>]"},
	{"a = true = d", "[<rule: 0-1-5 field: 0>]"},
	{"false = true = d", "[<rule: 0-1-5>]"},
	{"false = a = d", "[<rule: 0-1-5>]"},
	{"a = b; c d = e", "[<rule: 0-1-4 field: 0> <rule: 4-6-8 field: 5>]"},
	{";=", "[<rule: 0-0-1> <rule: 1-1-2>]"},
	{";;", "[<rule: 0-0-1> <rule: 1-1-2>]"},
	{"a = b", "[<rule: 0-1-3 field: 0>]"},
	{"a = b;", "[<rule: 0-1-4 field: 0>]"},
	{"a = b c d = d", "[<rule: 0-1-7 field: 0>]"},
	{"c d = d", "[<rule: 0-2-4 field: 1>]"},
	{"a = b c; d = d", "[<rule: 0-1-5 field: 0> <rule: 5-6-8 field: 5>]"},
	{"a ;; b", "[<rule: 0-1-2 field: 0> <rule: 2-2-3> <rule: 3-4-4 field: 3>]"},
	{"#\na#\n=#\nb#\n;#\nc#\nd#\n=#\nd#\n", "[<rule: 0-3-8 field: 1> <rule: 8-13-17 field: 11>]"},
}

func TestParser(t *testing.T) {
	for i, params := range fixture {
		p := NewParser(
			meta.Meta{"foo": meta.Int, "bar": meta.String},
			model.Set{"x": {}, "y": {}},
			model.Set{"baz": {}, "quux": {}})
		p.Parse(tokenizer.TokenizeString(params.rules))
		got := []Diagnostic{}

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
	{"abc = 1;", nil},
	{"foo = 1;# comment", nil},
	{"foo = 1", nil},
	{"foo = 1;foo = 2;", nil},
	{"foo = (baz 123);", nil},
	{"foo = (unknown 123);", nil},
	{"foo = = 1y;", nil},
	{"foo = 1m;;;", nil},
	{"bar = foo;", nil},
	{"foo = 1m;bar = foo;", nil},
	{"foo = $x; bar = $a;", nil},
	{"foo = (((;", nil},
}
