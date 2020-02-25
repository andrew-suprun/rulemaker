package main

import (
	"flag"
	"fmt"
	"os"

	"league.com/rulemaker/canonical_model"
	"league.com/rulemaker/content"
	"league.com/rulemaker/meta"
	"league.com/rulemaker/model"
	"league.com/rulemaker/style"
	"league.com/rulemaker/window"
)

var (
	lightFlag = flag.Bool("light", false, "Light theme")
	darkFlag  = flag.Bool("dark", false, "Dark theme")
)

func main() {
	flag.Parse()
	metainfo := meta.Metainfo(canonical_model.EmployeeDTO{})

	var inputs = model.Set{
		"policy":                        {},
		"sin":                           {},
		"employee_id":                   {},
		"last_name":                     {},
		"given_names":                   {},
		"person_type":                   {},
		"effective_date":                {},
		"transaction_date":              {},
		"division":                      {},
		"benefit_class":                 {},
		"administrative_class":          {},
		"retirement_date":               {},
		"termination_date":              {},
		"deceased_date":                 {},
		"birth_date":                    {},
		"gender":                        {},
		"language":                      {},
		"street":                        {},
		"city":                          {},
		"province_state":                {},
		"postal_zip_code":               {},
		"foreign_country":               {},
		"hire_date":                     {},
		"province_of_employment":        {},
		"province_of_residence":         {},
		"employee_smoker":               {},
		"business_location":             {},
		"cost_centre":                   {},
		"tax_exempt":                    {},
		"does_employee_have_dependants": {},
		"spouse_or_common_law_spouse":   {},
		"num_of_dependants":             {},
		"bank_transit_id":               {},
		"bank_number":                   {},
		"bank_account_number":           {},
		"earnings_amount":               {},
		"earnings_frequency":            {},
		"dependant_name_on_drug_card":   {},
		"revision_reason":               {},
		"created_by":                    {},
	}

	var operations = model.Set{
		"strip_prefix":        {},
		"strip_leading_zeros": {},
		"first_of":            {},
		"map":                 {},
		"select":              {},
		"all":                 {},
		"any":                 {},
		"one_of":              {},
		"join":                {},
		"+":                   {},
		"*":                   {},
		"=":                   {},
		"!=":                  {},
		"<":                   {},
		">":                   {},
		"<=":                  {},
		">=":                  {},
		"min":                 {},
		"max":                 {},
		"has":                 {},
		"first_of_month":      {},
		"weekly_hours":        {},
		"config":              {},
		"fail":                {},
		"log":                 {},
		"ticket":              {},
		"contains":            {},
		"skip":                {},
	}

	// c, e := content.NewContent("test.rules")
	c, e := content.NewFileContent("emp.rules")
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	theme := style.BlueTheme
	if *darkFlag {
		theme = style.DarkTheme
	} else if *lightFlag {
		theme = style.LightTheme
	}
	w, e := window.NewWindow(c, metainfo, inputs, operations, theme)
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	w.Run()
}
