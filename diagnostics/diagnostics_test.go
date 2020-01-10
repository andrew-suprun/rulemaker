package diagnostics

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"league.com/rulemaker/meta"

	"league.com/rulemaker/tokenizer"
)

func TestParser(t *testing.T) {
	for i, params := range fixture {
		got := ScanTokens(
			meta.Meta{"foo": meta.Int, "bar": meta.String},
			Set{"x": {}, "y": {}},
			Set{"baz": {}, "quux": {}},
			tokenizer.Tokenize(splitLines(params.rules)),
		)

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
	errors []Diagnostic
}{
	{"", nil},
	{"# comment", nil},
	{"abc = 1;", []Diagnostic{
		{0, 0, "Canonical model does not have field \"abc\""},
	}},
	{"foo = 1;# comment", nil},
	{"foo = 1", nil},
	{"foo=1;foo=2;", []Diagnostic{
		{0, 6, "Redefinition of \"foo\" previously defined at 1:1"},
	}},
	{"foo=(baz 123);", nil},
	{"foo=(unknown 123);", []Diagnostic{
		{0, 5, "Operation \"unknown\" is not defined"},
	}},
	{"foo = = 1y;", []Diagnostic{
		{0, 6, "Extra '='"},
	}},
	{"foo = 1m;;;", nil},
	{"bar=foo;", []Diagnostic{
		{0, 4, "Canonical field \"foo\" is not defined"},
	}},
	{"foo = 1m;bar=foo;", nil},
	{"foo = $x; bar = $a;", []Diagnostic{
		{0, 16, "Input field \"$a\" is not defined"},
	}},
	{"foo = (((;", []Diagnostic{
		{0, 6, "Unbalanced '('"},
		{0, 7, "Unbalanced '('"},
		{0, 8, "Unbalanced '('"},
	}},
}
