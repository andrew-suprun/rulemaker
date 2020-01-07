package diagnostics

import (
	"fmt"
	"sort"
	"strings"

	"league.com/rulemaker/meta"
	"league.com/rulemaker/tokenizer"
)

type Diagnostic struct {
	LineIndex, TokenIndex int
	Message               string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%d:%d: %s", d.LineIndex, d.TokenIndex, d.Message)
}

type Set map[string]struct{}

func ScanTokens(metainfo meta.Meta, inputs, operations Set, tokens tokenizer.Tokens) []Diagnostic {
	p := &scanner{
		metainfo:   metainfo,
		inputs:     inputs,
		operations: operations,
		tokens:     tokens,
		headers:    map[string]tokenizer.Token{},
	}
	p.scanTokens()

	sort.Slice(p.diagnostics, func(i, j int) bool {
		if p.diagnostics[i].LineIndex < p.diagnostics[j].LineIndex {
			return true
		}
		if p.diagnostics[i].LineIndex > p.diagnostics[j].LineIndex {
			return false
		}
		return p.diagnostics[i].TokenIndex < p.diagnostics[j].TokenIndex
	})

	return p.diagnostics
}

type tokenPosition struct {
	lineIndex, tokenIndex int
}

type scanner struct {
	metainfo          meta.Meta
	inputs            Set
	operations        Set
	tokens            tokenizer.Tokens
	headers           map[string]tokenizer.Token
	diagnostics       []Diagnostic
	state             state
	header            []tokenizer.Token
	body              []tokenizer.Token
	previousTokenType tokenizer.TokenType
	openParens        []tokenPosition
	lineIndex         int
	tokenIndex        int
	token             tokenizer.Token
}

type state int

const (
	expectHeader state = iota
	expectBody
)

func (s *scanner) scanTokens() {
	for s.lineIndex = range s.tokens {
		for s.tokenIndex, s.token = range s.tokens[s.lineIndex] {
			s.scanToken()
		}
	}
}

func (s *scanner) scanToken() {
	if s.token.Type == tokenizer.Comment {
		return
	}
	if s.token.Type == tokenizer.InvalidToken {
		s.report("Invalid Token")
		return
	}

	switch s.token.Type {
	case tokenizer.EqualSign:
		s.equalSign()
	case tokenizer.Semicolon:
		s.semicolon()
	case tokenizer.OpenParen:
		s.openParen()
	case tokenizer.CloseParen:
		s.closeParen()
	default:
		s.regularToken()
	}
	s.previousTokenType = s.token.Type
}

func (s *scanner) report(message string, args ...interface{}) {
	s.diagnostics = append(s.diagnostics, Diagnostic{
		LineIndex:  s.lineIndex,
		TokenIndex: s.tokenIndex,
		Message:    fmt.Sprintf(message, args...),
	})
}

func (s *scanner) regularToken() {
	if s.state == expectHeader {
		s.headerToken()
	} else if s.state == expectBody {
		s.bodyToken()
	}
}

func (s *scanner) headerToken() {
	if s.token.Type != tokenizer.Identifier && s.token.Type != tokenizer.Variable && s.token.Type != tokenizer.InvalidToken {
		s.report("Rule header must be either canonical field or temporary variable")
	}
	if len(s.header) > 0 {
		s.report("Rule header and body must be separated with '='")
	}
	if previousHeader, alreadyDefined := s.headers[s.token.Text]; alreadyDefined {
		s.report("Redefinition of %q previously defined at %d:%d",
			s.token.Text, 999, previousHeader.Column+1) // TODO: fix line
	}
	if s.token.Type == tokenizer.Identifier {
		if s.metainfo.Type(s.token.Text) == meta.Invalid {
			s.report("Canonical model does not have field %q", s.token.Text)
		}
	}
	s.headers[s.token.Text] = s.token
	s.header = append(s.header, s.token)
}

func (s *scanner) bodyToken() {
	if s.previousTokenType == tokenizer.OpenParen {
		if _, defined := s.operations[s.token.Text]; s.token.Type != tokenizer.Identifier || !defined {
			s.report("Operation %q is not defined", s.token.Text)
		}
	} else {
		if s.token.Type == tokenizer.Identifier {
			if s.metainfo.Type(s.token.Text) == meta.Invalid {
				s.report("Canonical model does not have field %q", s.token.Text)
			}
		}
		if s.token.Type == tokenizer.Identifier || s.token.Type == tokenizer.Variable {
			if _, defined := s.headers[s.token.Text]; !defined {
				s.report("Field %q is not defined", s.token.Text)
			}
		}
	}
	if s.token.Type == tokenizer.Input {
		input, _ := s.token.Value.(string)
		inputParts := strings.Split(input, ":")
		if _, defined := s.inputs[inputParts[0]]; !defined {
			s.report("Input field %q is not defined", s.token.Text)
		}
	}
	s.body = append(s.body, s.token)
}

func (s *scanner) equalSign() {
	if s.state == expectHeader {
		if len(s.header) == 0 {
			s.report("Rule header is missing")
		}
		s.state = expectBody
	} else {
		s.report("Extra '='")
	}
	s.header = append(s.header, s.token)
}

func (s *scanner) semicolon() {
	for _, p := range s.openParens {
		s.diagnostics = append(s.diagnostics, Diagnostic{
			LineIndex:  p.lineIndex,
			TokenIndex: p.tokenIndex,
			Message:    fmt.Sprintf("Unbalanced '('"),
		})
	}
	s.header = s.header[:0]
	s.body = s.body[:0]
	s.openParens = s.openParens[:0]
	s.state = expectHeader
}

func (s *scanner) openParen() {
	if s.state != expectBody {
		s.report("Unexpected '('")
		s.header = append(s.header, s.token)
		return
	}
	s.openParens = append(s.openParens, tokenPosition{lineIndex: s.lineIndex, tokenIndex: s.tokenIndex})
	s.body = append(s.body, s.token)
}

func (s *scanner) closeParen() {
	if s.state != expectBody {
		s.report("Unexpected ')'")
		s.header = append(s.header, s.token)
		return
	}
	if len(s.openParens) == 0 {
		s.report("Unbalanced ')'")
	} else {
		s.openParens = s.openParens[:len(s.openParens)-1]
	}
	s.body = append(s.body, s.token)
}
