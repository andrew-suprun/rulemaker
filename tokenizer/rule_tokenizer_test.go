package tokenizer

import (
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
)

type tokens struct {
	tokens []Token
}

func (t *tokens) Token(token Token) {
	t.tokens = append(t.tokens, token)
}

func (t *tokens) Done() {}

func TestLineTokenizer(t *testing.T) {
	for _, f := range fixture {
		listener := &tokens{}
		Tokenize([][]rune{[]rune(f.line)}, listener)
		if len(listener.tokens) == 0 {
			log.Println("line    ", f.line)
			log.Println("got no tokens")
			t.FailNow()
		}
		if !reflect.DeepEqual(listener.tokens, f.tokens) {
			log.Println("line    ", f.line)
			log.Println("got     ", listener.tokens)
			log.Println("expected", f.tokens)
			t.FailNow()
		}
	}
}

func TestTokensDontOverlap(t *testing.T) {
	Tokenize(rulesText, (&listener{t: t}))
}

var fixture = []struct {
	line   string
	tokens []Token
}{
	{`(foo bar)`, []Token{{OpenParen, 0, 0, 1, `(`, nil}, {Function, 0, 1, 4, `foo`, `foo`}, {CanonicalField, 0, 5, 8, `bar`, `bar`}, {CloseParen, 0, 8, 9, `)`, nil}}},
	{`"abc"`, []Token{{StringLiteral, 0, 0, 5, `"abc"`, `abc`}}},
	{`"abc`, []Token{{InvalidToken, 0, 0, 4, `"abc`, nil}}},
	{`  "\""`, []Token{{StringLiteral, 0, 2, 6, `"\""`, `"`}}},
	{`"\`, []Token{{InvalidToken, 0, 0, 2, `"\`, nil}}},
	{`""`, []Token{{StringLiteral, 0, 0, 2, `""`, ""}}},
	{`   # abc   `, []Token{{Comment, 0, 3, 11, `# abc   `, nil}}},
	{` (())`, []Token{{OpenParen, 0, 1, 2, `(`, nil}, {OpenParen, 0, 2, 3, `(`, nil}, {CloseParen, 0, 3, 4, `)`, nil}, {CloseParen, 0, 4, 5, `)`, nil}}},
	{` @2020-01-02 `, []Token{{DateLiteral, 0, 1, 12, `@2020-01-02`, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-02`, []Token{{DateLiteral, 0, 0, 11, `@2020-01-02`, time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}}},
	{`@2020-01-0 `, []Token{{InvalidToken, 0, 0, 10, `@2020-01-0`, nil}}},
	{`_`, []Token{{Variable, 0, 0, 1, `_`, `_`}}},
	{`  _abc_123_ `, []Token{{Variable, 0, 2, 11, `_abc_123_`, `_abc_123_`}}},
	{`abc`, []Token{{CanonicalField, 0, 0, 3, `abc`, `abc`}}},
	{` abc.+.-.123 `, []Token{{CanonicalField, 0, 1, 12, `abc.+.-.123`, `abc.+.-.123`}}},
	{` today `, []Token{{TodayLiteral, 0, 1, 6, `today`, nil}}},
	{`nil`, []Token{{NilLiteral, 0, 0, 3, `nil`, nil}}},
	{`true`, []Token{{BooleanLiteral, 0, 0, 4, `true`, true}}},
	{`false`, []Token{{BooleanLiteral, 0, 0, 5, `false`, false}}},
	{`  -00123  `, []Token{{IntegerLiteral, 0, 2, 8, `-00123`, -123}}},
	{`  -00123.  `, []Token{{FloatingPointLiteral, 0, 2, 9, `-00123.`, -123.0}}},
	{`1y 2m 3d`, []Token{{YearSpanLiteral, 0, 0, 2, `1y`, 1}, {MonthSpanLiteral, 0, 3, 5, `2m`, 2}, {DaySpanLiteral, 0, 6, 8, `3d`, 3}}},
	{`123 abc`, []Token{{IntegerLiteral, 0, 0, 3, `123`, 123}, {CanonicalField, 0, 4, 7, `abc`, `abc`}}},
	{`a = b;`, []Token{{CanonicalField, 0, 0, 1, `a`, `a`}, {EqualSign, 0, 2, 3, `=`, nil}, {CanonicalField, 0, 4, 5, `b`, `b`}, {Semicolon, 0, 5, 6, `;`, nil}}},
}

type listener struct {
	t      *testing.T
	line   int
	column int
}

func (l *listener) Token(token Token) {
	if token.Line < l.line || token.Line == l.line && token.StartColumn < l.column {
		l.t.FailNow()
	}
	l.line = token.Line
	l.column = token.EndColumn
}

func (l *listener) Done() {
}

func splitLines(text string) [][]rune {
	lines := strings.Split(text, "\n")
	result := make([][]rune, len(lines))
	for i, line := range lines {
		result[i] = []rune(line)
	}
	return result
}

var rulesText = splitLines(`_policy (map $policy
                from: "59450" to: "170150"
                from: "59470" to: "170170"
                from: "59471" to: "170170-PT"
                from: "59475" to: "170175"
                from: "59476" to: "170175-PT"
                from: "59455" to: "170155"
                from: "59465" to: "170165"
                from: "59460" to: "170160"
                from: "170150" to: "170150"
                from: "170170" to: "170170"
                from: "170175" to: "170175"
                from: "170155" to: "170155"
                from: "170165" to: "170165"
                from: "170160" to: "170160"
                else: $policy)

_foo (+ @1991-01-02 @1991-01-2 true false nil today abc _xyz)

_employee_id_without_prefix (strip_leading_zeros (strip_prefix $employee_id prefix: "0PT"))

_employee_id (select
            when: (contains $employee_id substring: "0PT") apply: (join "PT-" _employee_id_without_prefix)
            else: _employee_id_without_prefix)

# Benefit Class Lookup based on the "Policy-Benefit Class" combination
_policy_benefit_class (join $policy $benefit_class separator: "-")
_league_benefit_class (map _policy_benefit_class
        from: "170150-1" to: "170150 Class 1,2"
        from: "59476-31" to: "59475 Class 31"
        else: (fail "Invalid Benefit Class")
    )

# Mapping for ASO policies to be reviewed/confirmed. TODO : IN-2167
_employment_status (select
  when: (one_of _policy_benefit_class "170150-3" "170165-23" "170165-26" "170165-30" "170165-31" "170170-11" "170170-12" "59471-11" "59471-12" "59476-31")
  apply: "Part-Time"
  else: "Full-Time")

employee_id (join _policy _employee_id separator: "-")

last_name $last_name
first_name $given_names

state (select
  when: (has suspended_date) apply: "terminated"
  else: "active")
date_of_birth $birth_date:date

sex (map $gender 
  from: "F" to: "Female" 
  from: "M" to: "Male" 
  else: $gender)

locale (map $language 
  from: "F" to: "fr-CA" 
  from: "E" to: "en-CA" 
  else: $language)

address1 $street
city $city
province $province_state
postal_code $postal_zip_code
country (first_of $foreign_country "Canada")
annual_earnings 1.00

custom_fields.life_member_id _employee_id_without_prefix
custom_fields.life_policy_id $policy

# policy
# sin
`)
