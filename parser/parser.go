package parser

import (
	"fmt"
	"sort"
	"strings"

	"league.com/rulemaker/meta"
	"league.com/rulemaker/model"
	"league.com/rulemaker/tokenizer"
)

type Rule struct {
	Index int
	Head  int
	Body  int
	End   int
	Field int
}

type Rules []Rule

func (r Rule) String() string {
	if r.Field >= 0 {
		return fmt.Sprintf("<rule: %d-%d-%d field: %d>", r.Head, r.Body, r.End, r.Field)
	}
	return fmt.Sprintf("<rule: %d-%d-%d>", r.Head, r.Body, r.End)
}

type Diagnostic struct {
	Token   tokenizer.Token
	Message string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%d:%d: %s", d.Token.Line(), d.Token.StartColumn(), d.Message)
}

func NewParser(metainfo meta.Meta, inputs, operations model.Set) *Parser {
	return &Parser{
		metainfo:   metainfo,
		inputs:     inputs,
		operations: operations,
	}
}

type Parser struct {
	metainfo   meta.Meta
	inputs     model.Set
	operations model.Set

	tokens      tokenizer.Tokens
	rules       Rules
	diagnostics []Diagnostic
	completions []string
}

type Completion struct {
	Name      string
	TokenType tokenizer.TokenType
}

func (p *Parser) Parse(tokens tokenizer.Tokens) {
	p.tokens = tokens
	p.diagnostics = p.diagnostics[:0]
	p.makeRules()
	p.scanDefinitions()
	p.scanRules()
	p.sortDiagnostics()
}

func (p *Parser) makeRules() {
	p.rules = p.rules[:0]
	startIndex := 0
	head := true
	rule := Rule{Index: len(p.rules), Field: -1}
	for index, token := range p.tokens {
		if head {
			switch token.Type() {
			case tokenizer.EqualSign:
				rule.Head = startIndex
				rule.Body = index
				rule.End = index + 1
				startIndex = index
				head = false
			case tokenizer.Semicolon:
				rule.Head = startIndex
				rule.Body = index
				rule.End = index + 1
				startIndex = index + 1
				p.rules = append(p.rules, rule)
				rule = Rule{Index: len(p.rules), Head: index + 1, Field: -1}
			case tokenizer.CanonicalField, tokenizer.Variable:
				rule.Body = index + 1
				rule.End = index + 1
				rule.Field = index
			case tokenizer.EndMarker:
				if rule.End > 0 {
					p.rules = append(p.rules, rule)
				}
			default:
				rule.Body = index + 1
				rule.End = index + 1
			}
		} else {
			switch token.Type() {
			case tokenizer.Semicolon:
				rule.End = index + 1
				startIndex = index + 1
				head = true
				p.rules = append(p.rules, rule)
				rule = Rule{Index: len(p.rules), Head: index + 1, Field: -1}
			case tokenizer.EndMarker:
				if rule.End > 0 {
					p.rules = append(p.rules, rule)
				}
			default:
				rule.End = index + 1
			}
		}
	}
}

func (p *Parser) scanDefinitions() {
	definitions := map[string][]int{}
	for i := range p.rules {
		definitionIndex := p.rules[i].Field
		if definitionIndex >= 0 {
			token := p.tokens[definitionIndex]
			ruleIndices := definitions[p.tokens.Text(token)]
			ruleIndices = append(ruleIndices, i)
			definitions[p.tokens.Text(token)] = ruleIndices
		}
	}
	for field, ruleIndices := range definitions {
		if len(ruleIndices) > 1 {
			for _, ruleIndex := range ruleIndices {
				p.report(p.tokens[p.rules[ruleIndex].Field], "Multiple definitions of '%v'", field)
			}
		}
	}
}

func (p *Parser) scanRules() {
	for _, rule := range p.rules {
		p.scanRuleHead(rule)
		p.scanRuleBody(rule)
	}
}

func (p *Parser) scanRuleHead(rule Rule) {
	for tokenIndex := rule.Head; tokenIndex < rule.Body; tokenIndex++ {
		token := p.tokens[tokenIndex]
		if token.Type() == tokenizer.InvalidToken {
			p.report(token, "Invalid token '%v'", p.tokens.Text(token))
		} else if token.Type() != tokenizer.Comment && tokenIndex != rule.Field {
			p.report(token, "Unexpected token '%v'", p.tokens.Text(token))
		} else if token.Type() == tokenizer.CanonicalField {
			if p.metainfo.Type(p.tokens.Text(token)) == meta.Invalid {
				p.report(token, "Canonical model does not have field '%v'", p.tokens.Text(token))
			}
		}
	}
}

func (p *Parser) scanRuleBody(rule Rule) {
	if rule.Head == rule.Body || rule.Body == rule.End {
		token := p.tokens[rule.Head]
		if token.Type() != tokenizer.Comment {
			p.report(token, "Incomplete rule")
			return
		}
	}
	firstToken := p.tokens[rule.Body]
	if firstToken.Type() != tokenizer.EqualSign {
		p.report(firstToken, "Missing '='")
		return
	}
	var openParentheses tokenizer.Tokens
	bodyComplete := false
	for tokenIndex := rule.Body + 1; tokenIndex < rule.End; tokenIndex++ {
		token := p.tokens[tokenIndex]
		if token.Type() == tokenizer.Semicolon {
			continue
		}
		switch token.Type() {
		case tokenizer.CanonicalField:
			if p.metainfo.Type(p.tokens.Text(token)) == meta.Invalid {
				p.report(token, "Canonical model does not have field '%v'", p.tokens.Text(token))
			} else if index := p.firstDefinition(p.tokens.Text(token)); index < 0 && index >= rule.Index {
				p.report(token, "Canonical field '%v' is not defined", p.tokens.Text(token))
			} else if bodyComplete {
				p.report(token, "Extraneous token '%v'", p.tokens.Text(token))
			} else if len(openParentheses) == 0 {
				bodyComplete = true
			}
		case tokenizer.Variable:
			if index := p.firstDefinition(p.tokens.Text(token)); index < 0 && index >= rule.Index {
				p.report(token, "Variable '%v' is not defined", p.tokens.Text(token))
			} else if bodyComplete {
				p.report(token, "Extraneous token '%v'", p.tokens.Text(token))
			} else if len(openParentheses) == 0 {
				bodyComplete = true
			}
		case tokenizer.Operation:
			if _, defined := p.operations[p.tokens.Text(token)]; !defined {
				p.report(token, "Operation '%v' is not defined", p.tokens.Text(token))
			}
		case tokenizer.Input:
			input, _ := token.Value().(string)
			inputParts := strings.Split(input, ":")
			if _, defined := p.inputs[inputParts[0]]; !defined {
				p.report(token, "Input field '%v' is not defined", p.tokens.Text(token))
			} else if bodyComplete {
				p.report(token, "Extraneous token '%v'", p.tokens.Text(token))
			} else if len(openParentheses) == 0 {
				bodyComplete = true
			}
		case tokenizer.OpenParenthesis:
			openParentheses = append(openParentheses, token)
			var nextToken *tokenizer.Token
			for nextTokenIndex := tokenIndex + 1; nextTokenIndex < rule.End; nextTokenIndex++ {
				next := p.tokens[nextTokenIndex]
				if next.Type() == tokenizer.Comment {
					continue
				}
				nextToken = &next
				break
			}
			if nextToken == nil || nextToken.Type() == tokenizer.OpenParenthesis || nextToken.Type() == tokenizer.CloseParenthesis {
				p.report(token, "Missing operation")
			} else if nextToken.Type() != tokenizer.Operation {
				p.report(*nextToken, "Missing operation")
			}
		case tokenizer.CloseParenthesis:
			if len(openParentheses) == 0 {
				p.report(token, "Unbalanced ')'")
			} else {
				openParentheses = openParentheses[:len(openParentheses)-1]
				if len(openParentheses) == 0 {
					bodyComplete = true
				}
			}
		case tokenizer.EqualSign:
			p.report(token, "Unexpected '='")
		default:
			if token.Type() != tokenizer.Comment {
				if bodyComplete {
					p.report(token, "Extraneous token '%v'", p.tokens.Text(token))
				} else if len(openParentheses) == 0 {
					bodyComplete = true
				}
			}
		}
	}
	for _, openParenthesis := range openParentheses {
		p.report(openParenthesis, "Unbalanced '('")
	}
}

func (p *Parser) firstDefinition(name string) int {
	for _, rule := range p.rules {
		if rule.Field >= 0 {
			if p.tokens.Text(p.tokens[rule.Field]) == name {
				return rule.Index
			}
		}
	}
	return -1
}

func (p *Parser) report(token tokenizer.Token, message string, args ...interface{}) {
	for _, d := range p.diagnostics {
		if d.Token.Line() == token.Line() && d.Token.StartColumn() == token.StartColumn() {
			return
		}
	}
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Token:   token,
		Message: fmt.Sprintf(message, args...),
	})
}

func (p *Parser) sortDiagnostics() {
	sort.Slice(p.diagnostics, func(i, j int) bool {
		if p.diagnostics[i].Token.Line() < p.diagnostics[j].Token.Line() {
			return true
		}
		if p.diagnostics[i].Token.Line() > p.diagnostics[j].Token.Line() {
			return false
		}
		return p.diagnostics[i].Token.StartColumn() < p.diagnostics[j].Token.StartColumn()
	})
}

func (p *Parser) Diagnostics() []Diagnostic {
	return p.diagnostics
}

func (p *Parser) Completions(line, column int) []Completion {
	var rule Rule
	for _, rule = range p.rules {
		token := p.tokens[rule.End-1]
		if token.Line() < line || (token.Line() == line && token.EndColumn() <= column) {
			continue
		}
		return p.completionsForRule(rule, line, column)
	}
	return p.completionsForHead(len(p.rules), "")
}

func (p *Parser) completionsForRule(rule Rule, line, column int) (result []Completion) {
	token := p.tokens[rule.Head]
	for i := rule.Head + 1; i < rule.End; i++ {
		next := p.tokens[i]
		if next.Type() == tokenizer.Comment {
			continue
		}
		if next.After(line, column-1) {
			break
		}
		token = next
	}

	prefix := ""
	if line == token.Line() && column > token.StartColumn() && column <= -token.EndColumn() {
		prefix = p.tokens.Text(token)[:column-token.StartColumn()]
	}
	tokenType := token.Type()
	if !token.Contains(line, column-1) {
		tokenType = tokenizer.InvalidToken
	}
	if p.tokens[rule.Body].After(line, column-1) {
		result = p.completionsForHead(rule.Index, prefix)
	} else {
		if tokenType == tokenizer.OpenParenthesis {
			prefix = ""
		}
		result = p.completionsForBody(rule.Index, prefix, tokenType)
	}
	p.completions = make([]string, len(result))
	prefixLen := len(prefix)
	for i, completion := range result {
		if completion.Name[0] == ' ' {
			p.completions[i] = completion.Name[prefixLen+1:]
		}
		p.completions[i] = completion.Name[prefixLen:]
	}
	return result
}

func (p *Parser) completionsForHead(ruleIndex int, prefix string) []Completion {
	completions := map[string]tokenizer.TokenType{}
	for name := range p.metainfo {
		completions[name] = tokenizer.CanonicalField
	}
	for _, prevRule := range p.rules[:ruleIndex] {
		if prevRule.Field != -1 {
			prevToken := p.tokens[prevRule.Field]
			delete(completions, p.tokens.Text(prevToken))
		}
	}

	result := filterByPrefix(completions, prefix)
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func filterByPrefix(completions map[string]tokenizer.TokenType, prefix string) (result []Completion) {
	if prefix == "" || prefix == "=" {
		for name, tType := range completions {
			result = append(result, Completion{Name: name, TokenType: tType})
		}
	} else {
		for name, tType := range completions {
			if strings.HasPrefix(name, prefix) {
				result = append(result, Completion{Name: name, TokenType: tType})
			}
		}
	}
	return result
}

func (p *Parser) completionsForBody(ruleIndex int, prefix string, tokenType tokenizer.TokenType) (result []Completion) {
	completions := map[string]tokenizer.TokenType{}
	if tokenType == tokenizer.OpenParenthesis || tokenType == tokenizer.Operation {
		for op := range p.operations {
			completions[op] = tokenizer.Operation
		}
	} else { // TODO: implement other types
		for input := range p.inputs {
			completions["$"+input] = tokenizer.Input
		}
		for _, rule := range p.rules[:ruleIndex] {
			if rule.Field != -1 {
				token := p.tokens[rule.Field]
				text := p.tokens.Text(token)
				if text[0] == '_' {
					completions[text] = tokenizer.Variable
				} else {
					completions[text] = tokenizer.CanonicalField
				}
			}
		}
	}

	result = filterByPrefix(completions, prefix)
	sort.Slice(result, func(i, j int) bool {
		one := result[i].Name
		if one[0] == '$' || one[0] == '_' {
			one = one[1:] + one[:1]
		}
		two := result[j].Name
		if two[0] == '$' || two[0] == '_' {
			two = two[1:] + two[:1]
		}

		return one < two
	})
	return result
}

func (p *Parser) TotalCompletions() int {
	return len(p.completions)
}

func (p *Parser) Completion(lineNum int) string {
	if lineNum < len(p.completions) {
		return p.completions[lineNum]
	}
	return ""
}
