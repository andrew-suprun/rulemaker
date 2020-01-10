package diagnostics

import (
	"fmt"
	"sort"
	"strings"

	"league.com/rulemaker/meta"
	"league.com/rulemaker/tokenizer"
)

type Diagnostic struct {
	Line, Column int
	Message      string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%d:%d: %s", d.Line, d.Column, d.Message)
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
		if p.diagnostics[i].Line < p.diagnostics[j].Line {
			return true
		}
		if p.diagnostics[i].Line > p.diagnostics[j].Line {
			return false
		}
		return p.diagnostics[i].Column < p.diagnostics[j].Column
	})

	return p.diagnostics
}

type scanner struct {
	metainfo    meta.Meta
	inputs      Set
	operations  Set
	tokens      tokenizer.Tokens
	headers     map[string]tokenizer.Token
	diagnostics []Diagnostic
	state       state
	header      []tokenizer.Token
	body        []tokenizer.Token
	openParens  tokenizer.Tokens
	token       tokenizer.Token
}

type state int

const (
	expectHeader state = iota
	expectBody
)

func (s *scanner) scanTokens() {
	for _, s.token = range s.tokens {
		s.scanToken()
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
}

func (s *scanner) report(message string, args ...interface{}) {
	s.diagnostics = append(s.diagnostics, Diagnostic{
		Line:    s.token.Line,
		Column:  s.token.Column,
		Message: fmt.Sprintf(message, args...),
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
	if s.token.Type != tokenizer.CanonicalField && s.token.Type != tokenizer.Variable && s.token.Type != tokenizer.InvalidToken {
		s.report("Rule header must be either canonical field or temporary variable")
	}
	if len(s.header) > 0 {
		s.report("Rule header and body must be separated with '='")
	}
	if previousHeader, alreadyDefined := s.headers[s.token.Text]; alreadyDefined {
		s.report("Redefinition of %q previously defined at %d:%d",
			s.token.Text, previousHeader.Line+1, previousHeader.Column+1)
	}
	if s.token.Type == tokenizer.CanonicalField {
		if s.metainfo.Type(s.token.Text) == meta.Invalid {
			s.report("Canonical model does not have field %q", s.token.Text)
		}
	}
	s.headers[s.token.Text] = s.token
	s.header = append(s.header, s.token)
}

func (s *scanner) bodyToken() {
	if s.token.Type == tokenizer.Operation {
		if _, defined := s.operations[s.token.Text]; !defined {
			s.report("Operation %q is not defined", s.token.Text)
		}
	}
	if s.token.Type == tokenizer.CanonicalField {
		if s.metainfo.Type(s.token.Text) == meta.Invalid {
			s.report("Canonical model does not have field %q", s.token.Text)
		} else if _, defined := s.headers[s.token.Text]; !defined {
			s.report("Canonical field %q is not defined", s.token.Text)
		}
	}
	if s.token.Type == tokenizer.Variable {
		if _, defined := s.headers[s.token.Text]; !defined {
			s.report("Variable %q is not defined", s.token.Text)
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
			Line:    p.Line,
			Column:  p.Column,
			Message: fmt.Sprintf("Unbalanced '('"),
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
	s.openParens = append(s.openParens, s.token)
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
