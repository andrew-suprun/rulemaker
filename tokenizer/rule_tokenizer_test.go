package tokenizer

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
)

var fixture = []struct {
	line   string
	tokens []Token
}{
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
	{`abc`, []Token{{Identifier, 0, 0, 3, `abc`, `abc`}}},
	{` abc.+.-.123 `, []Token{{Identifier, 0, 1, 12, `abc.+.-.123`, `abc.+.-.123`}}},
	{` today `, []Token{{TodayLiteral, 0, 1, 6, `today`, nil}}},
	{`nil`, []Token{{NilLiteral, 0, 0, 3, `nil`, nil}}},
	{`true`, []Token{{BooleanLiteral, 0, 0, 4, `true`, true}}},
	{`false`, []Token{{BooleanLiteral, 0, 0, 5, `false`, false}}},
	{`  -00123  `, []Token{{IntegerLiteral, 0, 2, 8, `-00123`, -123}}},
	{`  -00123.  `, []Token{{FloatingPointLiteral, 0, 2, 9, `-00123.`, -123.0}}},
	{`1y 2m 3d`, []Token{{YearSpanLiteral, 0, 0, 2, `1y`, 1}, {MonthSpanLiteral, 0, 3, 5, `2m`, 2}, {DaySpanLiteral, 0, 6, 8, `3d`, 3}}},
	{`123 abc`, []Token{{IntegerLiteral, 0, 0, 3, `123`, 123}, {Identifier, 0, 4, 7, `abc`, `abc`}}},
}

func TestLineTokenizer(t *testing.T) {
	for _, f := range fixture {
		var tokens []Token
		Tokenize([][]rune{[]rune(f.line)}, func(token Token) {
			tokens = append(tokens, token)
		})
		if !reflect.DeepEqual(tokens, f.tokens) {
			log.Println("line    ", f.line)
			log.Println("got     ", tokens)
			log.Println("expected", f.tokens)
			t.FailNow()
		}
	}
}

func TestTokenizer(t *testing.T) {
	Tokenize(rules, listener())
}

func listener() Tokens {
	line, column := 0, 0
	return func(t Token) {
		fmt.Println(t)
		if t.Line < line || t.Line == line && t.StartColumn < column {
			panic("FUBAR")
		}
		line = t.Line
		column = t.EndColumn
	}
}

func splitLines(text string) [][]rune {
	lines := strings.Split(text, "\n")
	result := make([][]rune, len(lines))
	for i, line := range lines {
		result[i] = []rune(line)
	}
	return result
}

var rules = splitLines(`
_policy (map $policy
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



_employee_id_without_prefix (strip_leading_zeros (strip_prefix $employee_id prefix: "0PT"))

_employee_id (select
	when: (contains $employee_id substring: "0PT") apply: (join "PT-" _employee_id_without_prefix)
	else: _employee_id_without_prefix)

# Benefit Class Lookup based on the "Policy-Benefit Class" combination
_policy_benefit_class (join $policy $benefit_class separator: "-")
_league_benefit_class (map _policy_benefit_class
from: "170150-1" to: "170150 Class 1,2"
from: "170150-2" to: "170150 Class 1,2"
from: "170150-3" to: "170150 Class 3"
from: "170150-4" to: "170150 Class 4"
from: "170150-6" to: "170150 Class 6"
from: "170150-7" to: "170150 Class 7"
from: "170155-1" to: "170155 Class 1,2"
from: "170155-2" to: "170155 Class 1,2"
from: "170155-5" to: "170155 Class 5"
from: "170155-7" to: "170155 Class 7,11"
from: "170155-8" to: "170155 Class 8"
from: "170155-9" to: "170155 Class 9"
from: "170155-10" to: "170155 Class 10"
from: "170155-11" to: "170155 Class 7,11"
from: "170160-1" to: "170160 Class 1"
from: "170160-2" to: "170160 Class 2"
from: "170160-3" to: "170160 Class 3"
from: "170160-4" to: "170160 Class 4"
from: "170165-1" to: "170165 Class 1,2"
from: "170165-2" to: "170165 Class 1,2"
from: "170165-3" to: "170165 Class 3"
from: "170165-4" to: "170165 Class 4"
from: "170165-5" to: "170165 Class 5"
from: "170165-21" to: "170165 Class 21,22"
from: "170165-22" to: "170165 Class 21,22"
from: "170165-23" to: "170165 Class 23"
from: "170165-25" to: "170165 Class 25"
from: "170165-26" to: "170165 Class 26"
from: "170165-27" to: "170165 Class 27"
from: "170165-28" to: "170165 Class 28"
from: "170165-30" to: "170165 Class 30"
from: "170165-31" to: "170165 Class 31"
from: "170165-51" to: "170165 Class 51,52"
from: "170165-52" to: "170165 Class 51,52"
from: "170165-53" to: "170165 Class 53"
from: "170165-54" to: "170165 Class 54"
from: "170165-55" to: "170165 Class 55"
from: "170165-56" to: "170165 Class 56"
from: "170170-1" to: "170170 Class 1,2"
from: "170170-2" to: "170170 Class 1,2"
from: "170170-3" to: "170170 Class 3"
from: "170170-4" to: "170170 Class 4"
from: "170170-5" to: "170170 Class 5"
from: "170170-11" to: "170170 Class 11,12"
from: "170170-12" to: "170170 Class 11,12"
from: "170175-21" to: "170175 Class 21"
from: "170175-25" to: "170175 Class 25"
from: "170175-30" to: "170175 Class 30"
from: "170175-31" to: "170175 Class 31"

# Mapping for the 59* policies is yet to be confirmed/reviewed TODO : IN-2167
from: "59450-1" to: "59450 Class 1,2"
from: "59450-2" to: "59450 Class 1,2"
from: "59450-3" to: "59450 Class 3"
from: "59450-4" to: "59450 Class 4,5"
from: "59450-5" to: "59450 Class 5,5"
from: "59450-11" to: "59450 Class 11,12"
from: "59450-12" to: "59450 Class 11,12"
from: "59455-1" to: "59455 Class 1,2"
from: "59455-2" to: "59455 Class 1,2"
from: "59455-5" to: "59455 Class 5,6"
from: "59455-6" to: "59455 Class 6,6"
from: "59455-12" to: "59455 Class 12,13,14,15"
from: "59455-13" to: "59455 Class 12,13,14,15"
from: "59455-14" to: "59455 Class 12,13,14,15"
from: "59455-15" to: "59455 Class 12,13,14,15"
from: "59460-1" to: "59460 Class 1"
from: "59460-2" to: "59460 Class 2"
from: "59460-10" to: "59460 Class 10"
from: "59460-11" to: "59460 Class 11"
from: "59460-101" to: "59460 Class 101"
from: "59460-102" to: "59460 Class 102"
from: "59460-110" to: "59460 Class 110"
from: "59460-111" to: "59460 Class 111"
from: "59465-1" to: "59465 Class 1,2"
from: "59465-2" to: "59465 Class 1,2"
from: "59465-3" to: "59465 Class 3"
from: "59465-10" to: "59465 Class 10"
from: "59465-21" to: "59465 Class 21,22"
from: "59465-22" to: "59465 Class 21,22"
from: "59465-23" to: "59465 Class 23"
from: "59465-25" to: "59465 Class 25"
from: "59465-26" to: "59465 Class 26"
from: "59465-35" to: "59465 Class 35"
from: "59465-36" to: "59465 Class 36"
from: "59465-37" to: "59465 Class 37"
from: "59465-51" to: "59465 Class 51,52"
from: "59465-52" to: "59465 Class 52,52"
from: "59465-60" to: "59465 Class 60"
from: "59470-1" to: "59470 Class 1,2"
from: "59470-2" to: "59470 Class 1,2"
from: "59470-3" to: "59470 Class 3"
from: "59470-10" to: "59470 Class 10"
from: "59471-11" to: "59471 Class 11"
from: "59471-12" to: "59471 Class 12"
from: "59475-21" to: "59475 Class 21"
from: "59475-30" to: "59475 Class 30"
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
state_effective_date $effective_date:date
billing_division $division
benefit_class _league_benefit_class
suspended_date $termination_date:date

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
date_of_hire $hire_date:date
province_of_employment $province_of_employment
annual_earnings_effective_date $hire_date # ???
occupation "Employee" # IN-2143

email "example@example.com" # ???
employment_status _employment_status
annual_earnings 1.00

custom_fields.life_member_id _employee_id_without_prefix
custom_fields.life_policy_id $policy

# policy
# sin
# employee_id
# last_name
# given_names
# person_type
# effective_date
# transaction_date
# division
# benefit_class
# administrative_class
# retirement_date
# termination_date
# deceased_date
# birth_date
# gender
# language
# street
# city
# province_state
# postal_zip_code
# foreign_country
# hire_date
# province_of_employment
# province_of_residence
# employee_smoker
# business_location
# cost_centre
# tax_exempt
# does_employee_have_dependants
# spouse_or_common_law_spouse
# num_of_dependants
# bank_transit_id
# bank_number
# bank_account_number
# earnings_amount
# earnings_frequency
# dependant_name_on_drug_card
# revision_reason
# created_by
`)
