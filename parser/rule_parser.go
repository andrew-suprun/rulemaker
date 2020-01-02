package parser

import (
	"fmt"
	"strings"

	"league.com/rulemaker/meta"
	"league.com/rulemaker/tokenizer"
)

type Tokens interface {
	Token(token ParsedToken)
	Done()
}

type ParsedToken struct {
	tokenizer.Token
	Diagnostic string
}

func (t ParsedToken) String() string {
	if t.Diagnostic == "" {
		return fmt.Sprintf("<parsed token: %s>", t.Token)
	}
	return fmt.Sprintf("<parsed token: %s; diagnostic: %s>", t.Token, t.Diagnostic)
}

type Set map[string]struct{}

func Parse(content [][]rune, metainfo meta.Meta, inputs, operations Set, tokens Tokens) {
	tokenizer.Tokenize(content, newParser(metainfo, inputs, operations, tokens))
}

func newParser(metainfo meta.Meta, inputs, operations Set, tokens Tokens) *parser {
	return &parser{
		metainfo:   metainfo,
		inputs:     inputs,
		operations: operations,
		tokens:     tokens,
		headers:    map[string]ParsedToken{},
	}
}

type parser struct {
	metainfo    meta.Meta
	inputs      Set
	operations  Set
	headers     map[string]ParsedToken
	tokens      Tokens
	state       state
	header      []ParsedToken
	body        []ParsedToken
	openIndices []int
}

type state int

const (
	expectHeader state = iota
	expectBody
)

func (p *parser) Token(t tokenizer.Token) {
	token := parseToken(t)
	switch token.Type {
	case tokenizer.EqualSign:
		p.equalSign(token)
	case tokenizer.Semicolon:
		p.semicolon(token)
	case tokenizer.OpenParen:
		p.openParen(token)
	case tokenizer.CloseParen:
		p.closeParen(token)
	default:
		p.token(token)
	}
}

func (p *parser) token(token ParsedToken) {
	if p.state == expectHeader {
		p.headerToken(token)
	} else if p.state == expectBody {
		p.bodyToken(token)
	}
}

func (p *parser) headerToken(token ParsedToken) {
	if token.Type == tokenizer.Comment {
		p.tokens.Token(token)
		return
	}
	if token.Type != tokenizer.CanonicalField && token.Type != tokenizer.Variable && token.Type != tokenizer.InvalidToken {
		token.Diagnostic = "Rule header must be either canonical field or temporary variable"
	}
	if len(p.header) > 0 {
		token.Diagnostic = "Rule header and body must be separated with '='"
	}
	if previousHeader, alreadyDefined := p.headers[token.Text]; alreadyDefined {
		token.Diagnostic = fmt.Sprintf("Redefinition of %q; previously defined at %d:%d",
			token.Text, previousHeader.Line+1, previousHeader.StartColumn+1)
	}
	if token.Type == tokenizer.CanonicalField {
		if p.metainfo.Type(token.Text) == meta.Invalid {
			token.Diagnostic = fmt.Sprintf("Canonical model does not have field %q", token.Text)
		}
	}
	p.headers[token.Text] = token
	p.header = append(p.header, token)
}

func (p *parser) bodyToken(token ParsedToken) {
	if token.Type == tokenizer.CanonicalField {
		if p.metainfo.Type(token.Text) == meta.Invalid {
			token.Diagnostic = fmt.Sprintf("Canonical model does not have field %q", token.Text)
		}
	}
	if token.Type == tokenizer.CanonicalField || token.Type == tokenizer.Variable {
		if _, defined := p.headers[token.Text]; !defined {
			token.Diagnostic = fmt.Sprintf("Field %q is not defined", token.Text)
		}
	}
	if token.Type == tokenizer.Function {
		if _, defined := p.operations[token.Text]; !defined {
			token.Diagnostic = fmt.Sprintf("Operation %q is not defined", token.Text)
		}
	}
	if token.Type == tokenizer.Input {
		input, _ := token.Value.(string)
		inputParts := strings.Split(input, ":")
		if _, defined := p.inputs[inputParts[0]]; !defined {
			token.Diagnostic = fmt.Sprintf("Input %q is not defined", token.Text)
		}
	}
	p.body = append(p.body, token)
}

func (p *parser) equalSign(token ParsedToken) {
	if p.state == expectHeader {
		if len(p.header) == 0 {
			token.Diagnostic = "Rule header is missing"
		}
		p.state = expectBody
	} else {
		token.Diagnostic = "Extra '='"
	}
	p.header = append(p.header, token)
}

func (p *parser) semicolon(token ParsedToken) {
	for _, index := range p.openIndices {
		p.body[index].Diagnostic = "Unbalanced '('"
	}
	for _, t := range p.header {
		p.tokens.Token(t)
	}
	for _, t := range p.body {
		p.tokens.Token(t)
	}
	p.tokens.Token(token)
	p.header = p.header[:0]
	p.body = p.body[:0]
	p.openIndices = p.openIndices[:0]
	p.state = expectHeader
}

func (p *parser) openParen(token ParsedToken) {
	if p.state != expectBody {
		token.Diagnostic = "Unexpected '('"
		p.header = append(p.header, token)
		return
	}
	p.openIndices = append(p.openIndices, len(p.body))
	p.body = append(p.body, token)
}

func (p *parser) closeParen(token ParsedToken) {
	if p.state != expectBody {
		token.Diagnostic = "Unexpected ')'"
		p.header = append(p.header, token)
		return
	}
	if len(p.openIndices) == 0 {
		token.Diagnostic = "Unbalanced ')'"
	} else {
		p.openIndices = p.openIndices[:len(p.openIndices)-1]
	}
	p.body = append(p.body, token)
}

func parseToken(token tokenizer.Token) ParsedToken {
	diagnostic := ""
	if token.Type == tokenizer.InvalidToken {
		diagnostic = "Invalid Token"
	}
	return ParsedToken{Token: token, Diagnostic: diagnostic}
}

func (p *parser) Done() {
	for _, t := range p.header {
		p.tokens.Token(t)
	}
	for _, t := range p.body {
		p.tokens.Token(t)
	}
	p.tokens.Done()
}
