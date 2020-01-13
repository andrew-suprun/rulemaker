package parser

import (
	"fmt"
	"sort"
	"strings"

	"league.com/rulemaker/meta"
	"league.com/rulemaker/model"
	"league.com/rulemaker/tokenizer"
)

type Parser interface {
	Parse(tokens tokenizer.Tokens)
	Diagnostics() []Diagnostic
	Completions(line, column int) []string
	Completion(lineNum int) string
}

type Diagnostic struct {
	Line, Column int
	Message      string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%d:%d: %s", d.Line, d.Column, d.Message)
}

func NewParser(metainfo meta.Meta, inputs, operations model.Set) Parser {
	return &parser{
		metainfo:   metainfo,
		inputs:     inputs,
		operations: operations,
	}
}

type parser struct {
	metainfo   meta.Meta
	inputs     model.Set
	operations model.Set

	tokens      tokenizer.Tokens
	ruleStarts  []int
	diagnostics []Diagnostic
	headers     map[string]tokenizer.Token
	openParens  tokenizer.Tokens

	completions []string
	prefix      string
}

func (p *parser) Parse(tokens tokenizer.Tokens) {
	p.tokens = tokens
	p.diagnostics = p.diagnostics[:0]
	p.headers = map[string]tokenizer.Token{}
	p.splitRules()
	p.scanRules()
	p.sortDiagnostics()
}

func (p *parser) splitRules() {
	p.ruleStarts = p.ruleStarts[:0]
	nextStart := 0
	p.ruleStarts = append(p.ruleStarts, nextStart)
	for index, token := range p.tokens {
		if token.Type == tokenizer.Semicolon {
			p.splitRule(nextStart, index+1)
			nextStart = index + 1
			p.ruleStarts = append(p.ruleStarts, nextStart)
		}
	}
	if nextStart != len(p.tokens) {
		p.splitRule(nextStart, len(p.tokens))
		p.ruleStarts = append(p.ruleStarts, len(p.tokens))
	}
}

func (p *parser) splitRule(start, end int) {
	index := start + 1
	fieldIndex := -1
	for ; index < end; index++ {
		if p.tokens[index].Type == tokenizer.CanonicalField || p.tokens[index].Type == tokenizer.Variable {
			fieldIndex = index
		} else if p.tokens[index].Type == tokenizer.EqualSign && fieldIndex != -1 {
			p.ruleStarts = append(p.ruleStarts, fieldIndex)
			fieldIndex = -1
		} else if p.tokens[index].Type != tokenizer.Comment {
			fieldIndex = -1
		}
	}
}

func (p *parser) scanRules() {
	for i := range p.ruleStarts[:len(p.ruleStarts)-1] {
		p.scanRule(p.tokens[p.ruleStarts[i]:p.ruleStarts[i+1]])
	}
}

func (p *parser) scanRule(rule tokenizer.Tokens) {
	p.openParens = p.openParens[:0]
	for i, token := range rule {
		if token.Type == tokenizer.EqualSign {
			p.scanRuleHeader(rule[:i])
			p.scanRuleBody(rule[i+1:])
			for _, token := range p.openParens {
				p.report(token, "Unbalanced '('")
			}
			return
		} else if token.Type == tokenizer.InvalidToken {
			p.report(token, "Invalid token %s", token.Text)
			return
		}
	}
	for _, token := range rule {
		if token.Type != tokenizer.Comment && token.Type != tokenizer.Semicolon {
			p.report(token, "Incomplete rule")
			return
		}
	}
}

func (p *parser) scanRuleHeader(header tokenizer.Tokens) {
	lastFieldIndex := -1
	for i, token := range header {
		switch token.Type {
		case tokenizer.CanonicalField:
			if lastFieldIndex != -1 {
				p.report(header[i], "Unexpected canonical field %q", token.Text)
			} else {
				if p.metainfo.Type(token.Text) == meta.Invalid {
					p.report(token, "Canonical model does not have field %q", token.Text)
				} else if previousHeader, alreadyDefined := p.headers[token.Text]; alreadyDefined {
					p.report(token, "Canonical field %q redefined; previously defined at %d:%d",
						token.Text, previousHeader.Line+1, previousHeader.Column+1)
				}
				p.headers[token.Text] = token
			}
			lastFieldIndex = i
		case tokenizer.Variable:
			if lastFieldIndex != -1 {
				p.report(header[i], "Unexpected variable %q", token.Text)
			} else if previousHeader, alreadyDefined := p.headers[token.Text]; alreadyDefined {
				p.report(token, "Variable %q redefined; previously defined at %d:%d",
					token.Text, previousHeader.Line+1, previousHeader.Column+1)
			}
			p.headers[token.Text] = token
			lastFieldIndex = i
		case tokenizer.Comment:
			// comments don't do anything
		default:
			p.report(header[i], "Unexpected token %q", token.Text)
		}
	}
}

func (p *parser) scanRuleBody(body tokenizer.Tokens) {
	for i, token := range body {
		switch token.Type {
		case tokenizer.CanonicalField:
			if p.metainfo.Type(token.Text) == meta.Invalid {
				p.report(token, "Canonical model does not have field %q", token.Text)
			} else if _, defined := p.headers[token.Text]; !defined {
				p.report(token, "Canonical field %q is not defined", token.Text)
			}
		case tokenizer.Variable:
			if _, defined := p.headers[token.Text]; !defined {
				p.report(token, "Variable %q is not defined", token.Text)
			}
		case tokenizer.Operation:
			if _, defined := p.operations[token.Text]; !defined {
				p.report(token, "Operation %q is not defined", token.Text)
			}
		case tokenizer.Input:
			input, _ := token.Value.(string)
			inputParts := strings.Split(input, ":")
			if _, defined := p.inputs[inputParts[0]]; !defined {
				p.report(token, "Input field %q is not defined", token.Text)
			}
		case tokenizer.OpenParen:
			p.openParens = append(p.openParens, token)
			var nextToken *tokenizer.Token
			for _, next := range body[i+1:] {
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
			if len(p.openParens) == 0 {
				p.report(token, "Unbalanced ')'")
			} else {
				p.openParens = p.openParens[:len(p.openParens)-1]
			}
		case tokenizer.EqualSign:
			p.report(token, "Unexpected '='")
		}
	}
}

func (p *parser) sortDiagnostics() {
	sort.Slice(p.diagnostics, func(i, j int) bool {
		if p.diagnostics[i].Line < p.diagnostics[j].Line {
			return true
		}
		if p.diagnostics[i].Line > p.diagnostics[j].Line {
			return false
		}
		return p.diagnostics[i].Column < p.diagnostics[j].Column
	})
}

func (p *parser) report(token tokenizer.Token, message string, args ...interface{}) {
	for _, d := range p.diagnostics {
		if d.Line == token.Line && d.Column == token.Column {
			return
		}
	}
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Line:    token.Line,
		Column:  token.Column,
		Message: fmt.Sprintf(message, args...),
	})
}

func (p *parser) Diagnostics() []Diagnostic {
	return p.diagnostics
}

func (p *parser) Completions(column, line int) []string {
	prefix, expectedType, header := "", tokenizer.CanonicalField, true
	rule, found := p.findRule(column, line)
	if !found {
		return []string{}
	}
	prefix, expectedType, header = p.findPrefix(column, line, rule)

	set := model.Set{}
	if expectedType == tokenizer.Operation {
		for op := range p.operations {
			set[op] = struct{}{}
		}
	} else {
		switch expectedType {
		case tokenizer.CanonicalField, tokenizer.Variable:
			if header {
				for name := range p.metainfo {
					set[name] = struct{}{}
				}
				for name := range p.headers {
					delete(set, name)
				}
			} else {
				ruleStartColumn := rule[0].Column
				ruleStartLine := rule[0].Line
				for name, token := range p.headers {
					if token.Line < ruleStartLine || (token.Line == ruleStartLine && token.Column < ruleStartColumn) {
						if strings.HasPrefix(prefix, "_") || p.metainfo.Type(token.Text) != meta.Invalid {
							set[name] = struct{}{}
						}
					}
				}
			}
		case tokenizer.Input:
			for input := range p.inputs {
				set["$"+input] = struct{}{}
			}
		}
	}
	if len(prefix) > 0 {
		for name := range set {
			if !strings.HasPrefix(name, prefix) {
				delete(set, name)
			}
		}
	}

	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	p.completions = names
	p.prefix = prefix
	return names
}

func (p *parser) findRule(column, line int) (rule tokenizer.Tokens, found bool) {
	i := 1
	for ; i < len(p.ruleStarts)-1; i++ {
		token := p.tokens[p.ruleStarts[i]]
		if line > token.Line || (line == token.Line && column >= token.Column) {
			continue
		}
		return p.tokens[p.ruleStarts[i-1]:p.ruleStarts[i]], true
	}
	return tokenizer.Tokens{}, false
}

func (p *parser) findPrefix(column, line int, rule tokenizer.Tokens) (prefix string, expectedType tokenizer.TokenType, inHeader bool) {
	index := 0
	for index = range rule {
		if rule[index].Type == tokenizer.EqualSign {
			break
		}
	}
	header := rule[:index]
	body := rule[index+1:]
	if line < rule[index].Line || (line == rule[index].Line && column <= rule[index].Column) {
		prefix, expectedType = p.findPrefixInHeader(column, line, header)
		return prefix, expectedType, true
	}
	prefix, expectedType = p.findPrefixInBody(column, line, body)
	return prefix, expectedType, false
}

func (p *parser) findPrefixInHeader(column, line int, header tokenizer.Tokens) (prefix string, expectedType tokenizer.TokenType) {
	for _, token := range header {
		if token.Type == tokenizer.Comment {
			continue
		}
		if line < token.Line || (line == token.Line && column <= token.Column) {
			return "", tokenizer.CanonicalField
		}
		if token.Type != tokenizer.CanonicalField && token.Type != tokenizer.Variable {
			return "", tokenizer.InvalidToken
		}
		if line == token.Line && column > token.Column && column <= token.Column+len(token.Text) {
			return token.Text[:column-token.Column], token.Type
		}
		break
	}
	return "", tokenizer.InvalidToken
}

func (p *parser) findPrefixInBody(column, line int, body tokenizer.Tokens) (prefix string, expectedType tokenizer.TokenType) {
	for _, token := range body {
		if line < token.Line || (line == token.Line && column <= token.Column) {
			break
		}
		if line > token.Line || (line == token.Line && column > token.Column+len(token.Text)) {
			continue
		}
		if token.Type == tokenizer.OpenParen {
			return "", tokenizer.Operation
		}
		if line == token.Line && column > token.Column {
			return token.Text[:column-token.Column], token.Type
		}
	}
	return "", tokenizer.CanonicalField
}

func (p *parser) Completion(lineNum int) string {
	if lineNum < len(p.completions) {
		return p.completions[lineNum][len(p.prefix):]
	}
	return ""
}
