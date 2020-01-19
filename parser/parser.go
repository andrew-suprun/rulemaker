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
	Index int32
	Head  int32
	Body  int32
	End   int32
	Field int32
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
	return fmt.Sprintf("%d:%d: %s", d.Token.Line, d.Token.Column, d.Message)
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
	rule := Rule{Index: int32(len(p.rules)), Field: -1}
	for index, token := range p.tokens {
		if head {
			switch token.Type {
			case tokenizer.EqualSign:
				rule.Head = int32(startIndex)
				rule.Body = int32(index)
				rule.End = int32(index + 1)
				startIndex = index
				head = false
			case tokenizer.Semicolon:
				rule.Head = int32(startIndex)
				rule.Body = int32(index)
				rule.End = int32(index + 1)
				startIndex = index + 1
				p.rules = append(p.rules, rule)
				rule = Rule{Index: int32(len(p.rules)), Head: int32(index + 1), Field: -1}
			case tokenizer.CanonicalField, tokenizer.Variable:
				rule.Body = int32(index + 1)
				rule.End = int32(index + 1)
				rule.Field = int32(index)
			case tokenizer.EndMarker:
				if rule.End > 0 {
					p.rules = append(p.rules, rule)
				}
			default:
				rule.Body = int32(index + 1)
				rule.End = int32(index + 1)
			}
		} else {
			switch token.Type {
			case tokenizer.Semicolon:
				rule.End = int32(index + 1)
				startIndex = index + 1
				head = true
				p.rules = append(p.rules, rule)
				rule = Rule{Index: int32(len(p.rules)), Head: int32(index + 1), Field: -1}
			case tokenizer.EndMarker:
				if rule.End > 0 {
					p.rules = append(p.rules, rule)
				}
			default:
				rule.End = int32(index + 1)
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
			ruleIndices := definitions[token.Text]
			ruleIndices = append(ruleIndices, i)
			definitions[token.Text] = ruleIndices
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
		if token.Type == tokenizer.InvalidToken {
			p.report(token, "Invalid token '%v'", token.Text)
		} else if token.Type != tokenizer.Comment && tokenIndex != rule.Field {
			p.report(token, "Unexpected token '%v'", token.Text)
		} else if token.Type == tokenizer.CanonicalField {
			if p.metainfo.Type(token.Text) == meta.Invalid {
				p.report(token, "Canonical model does not have field '%v'", token.Text)
			}
		}
	}
}

func (p *Parser) scanRuleBody(rule Rule) {
	if rule.Head == rule.Body || rule.Body == rule.End {
		token := p.tokens[rule.Head]
		if token.Type != tokenizer.Comment {
			p.report(token, "Incomplete rule")
			return
		}
	}
	firstToken := p.tokens[rule.Body]
	if firstToken.Type != tokenizer.EqualSign {
		p.report(firstToken, "Missing '='")
		return
	}
	var openParens tokenizer.Tokens
	bodyComplete := false
	for tokenIndex := rule.Body + 1; tokenIndex < rule.End; tokenIndex++ {
		token := p.tokens[tokenIndex]
		if token.Type == tokenizer.Semicolon {
			continue
		}
		switch token.Type {
		case tokenizer.CanonicalField:
			if p.metainfo.Type(token.Text) == meta.Invalid {
				p.report(token, "Canonical model does not have field '%v'", token.Text)
			} else if index := p.firstDefinition(token.Text); index < 0 && index >= rule.Index {
				p.report(token, "Canonical field '%v' is not defined", token.Text)
			} else if bodyComplete {
				p.report(token, "Extraneous token '%v'", token.Text)
			} else if len(openParens) == 0 {
				bodyComplete = true
			}
		case tokenizer.Variable:
			if index := p.firstDefinition(token.Text); index < 0 && index >= rule.Index {
				p.report(token, "Variable '%v' is not defined", token.Text)
			} else if bodyComplete {
				p.report(token, "Extraneous token '%v'", token.Text)
			} else if len(openParens) == 0 {
				bodyComplete = true
			}
		case tokenizer.Operation:
			if _, defined := p.operations[token.Text]; !defined {
				p.report(token, "Operation '%v' is not defined", token.Text)
			}
		case tokenizer.Input:
			input, _ := token.Value.(string)
			inputParts := strings.Split(input, ":")
			if _, defined := p.inputs[inputParts[0]]; !defined {
				p.report(token, "Input field '%v' is not defined", token.Text)
			} else if bodyComplete {
				p.report(token, "Extraneous token '%v'", token.Text)
			} else if len(openParens) == 0 {
				bodyComplete = true
			}
		case tokenizer.OpenParen:
			openParens = append(openParens, token)
			var nextToken *tokenizer.Token
			for nextTokenIndex := tokenIndex + 1; nextTokenIndex < rule.End; nextTokenIndex++ {
				next := p.tokens[nextTokenIndex]
				if next.Type == tokenizer.Comment {
					continue
				}
				nextToken = &next
				break
			}
			if nextToken == nil || nextToken.Type == tokenizer.OpenParen || nextToken.Type == tokenizer.CloseParen {
				p.report(token, "Missing operation")
			} else if nextToken.Type != tokenizer.Operation {
				p.report(*nextToken, "Missing operation")
			}
		case tokenizer.CloseParen:
			if len(openParens) == 0 {
				p.report(token, "Unbalanced ')'")
			} else {
				openParens = openParens[:len(openParens)-1]
				if len(openParens) == 0 {
					bodyComplete = true
				}
			}
		case tokenizer.EqualSign:
			p.report(token, "Unexpected '='")
		default:
			if token.Type != tokenizer.Comment {
				if bodyComplete {
					p.report(token, "Extraneous token '%v'", token.Text)
				} else if len(openParens) == 0 {
					bodyComplete = true
				}
			}
		}
	}
	for _, openParen := range openParens {
		p.report(openParen, "Unbalanced '('")
	}
}

func (p *Parser) firstDefinition(name string) int32 {
	for _, rule := range p.rules {
		if rule.Field >= 0 {
			if p.tokens[rule.Field].Text == name {
				return rule.Index
			}
		}
	}
	return -1
}

func (p *Parser) report(token tokenizer.Token, message string, args ...interface{}) {
	for _, d := range p.diagnostics {
		if d.Token.Line == token.Line && d.Token.Column == token.Column {
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
		if p.diagnostics[i].Token.Line < p.diagnostics[j].Token.Line {
			return true
		}
		if p.diagnostics[i].Token.Line > p.diagnostics[j].Token.Line {
			return false
		}
		return p.diagnostics[i].Token.Column < p.diagnostics[j].Token.Column
	})
}

func (p *Parser) Diagnostics() []Diagnostic {
	return p.diagnostics
}

func (p *Parser) Completions(line, column int) []Completion {
	var rule Rule
	for _, rule = range p.rules {
		token := p.tokens[rule.End-1]
		if token.Line < line || (token.Line == line && token.Column+len(token.Text) <= column) {
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
		if next.Type == tokenizer.Comment {
			continue
		}
		if next.After(line, column-1) {
			break
		}
		token = next
	}

	prefix := ""
	if column > token.Column && column <= token.Column+len(token.Text) {
		prefix = token.Text[:column-token.Column]
	}
	if !token.Contains(line, column-1) {
		token.Type = tokenizer.InvalidToken
	}
	if p.tokens[rule.Body].After(line, column-1) {
		result = p.completionsForHead(int(rule.Index), prefix)
	} else {
		if token.Type == tokenizer.OpenParen {
			prefix = ""
		}
		result = p.completionsForBody(int(rule.Index), prefix, token.Type)
	}
	p.completions = make([]string, len(result))
	prefixLen := len(prefix)
	for i, completion := range result {
		if strings.HasPrefix(completion.Name, " ") {
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
			delete(completions, prevToken.Text)
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
	if tokenType == tokenizer.OpenParen || tokenType == tokenizer.Operation {
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
				text := token.Text
				if strings.HasPrefix(text, "_") {
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

func (p *Parser) Completion(lineNum int) string {
	if lineNum < len(p.completions) {
		return p.completions[lineNum]
	}
	return ""
}
