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
}

func (p *parser) Parse(tokens tokenizer.Tokens) {
	p.tokens = tokens
	p.ruleStarts = p.ruleStarts[:0]
	p.diagnostics = p.diagnostics[:0]
	p.headers = map[string]tokenizer.Token{}
	p.splitRules()
	p.scanRules()
	p.sortDiagnostics()
}

func (p *parser) splitRules() {
	start := 0
	for index, token := range p.tokens {
		if token.Type == tokenizer.Semicolon {
			p.splitRulesByEqualSign(start, index+1)
			start = index + 1
		}
	}
	p.splitRulesByEqualSign(start, len(p.tokens))
	p.ruleStarts = append(p.ruleStarts, len(p.tokens))
}

type splitState int

const (
	expectFirstField splitState = iota
	expectFirstEquals
	expectField
	expectEquals
)

func (p *parser) splitRulesByEqualSign(firstIndex, lastIndex int) {
	if firstIndex == lastIndex {
		return
	}
	p.ruleStarts = append(p.ruleStarts, firstIndex)
	state := expectFirstField
	for index := firstIndex; index < lastIndex; index++ {
		tokenType := p.tokens[index].Type
		switch state {
		case expectFirstField:
			if tokenType == tokenizer.CanonicalField || tokenType == tokenizer.Variable {
				state = expectFirstEquals
			}
		case expectFirstEquals:
			if tokenType == tokenizer.EqualSign {
				state = expectField
			} else {
				state = expectFirstField
			}
		case expectField:
			if tokenType == tokenizer.CanonicalField || tokenType == tokenizer.Variable {
				state = expectEquals
			}
		case expectEquals:
			if tokenType == tokenizer.EqualSign {
				p.ruleStarts = append(p.ruleStarts, index-1)
				state = expectField
			} else if tokenType != tokenizer.CanonicalField && tokenType != tokenizer.Variable {
				state = expectField
			}
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
			p.report(rule[0], "Incomplete rule")
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

func (p *parser) Completions(line, column int) []string {
	var header bool
	var prefix string
	var tokenType tokenizer.TokenType
outer:
	for i := range p.ruleStarts[:len(p.ruleStarts)-1] {
		header = true
		for _, token := range p.tokens[p.ruleStarts[i]:p.ruleStarts[i+1]] {
			if token.Type == tokenizer.EqualSign {
				header = false
			}
			if line == token.Line && column >= token.Column && column <= token.Column+len(token.Text) {
				if token.Type == tokenizer.EqualSign && column == token.Column {
					header = true
				}
				prefix = token.Text[:column-token.Column]
				tokenType = token.Type
				break outer
			}
			if line < token.Line || (line == token.Line && column < token.Column) {
				break outer
			}
		}
	}
	if prefix == "" {
		tokenType = tokenizer.CanonicalField
	}

	switch tokenType {
	case tokenizer.CanonicalField:
		set := model.Set{}
		for name := range p.metainfo {
			set[name] = struct{}{}
		}
		if header {
			for name := range p.headers {
				delete(set, name)
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
		return names
	}

	return nil
}
