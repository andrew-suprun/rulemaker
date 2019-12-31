package rule_parser

type Rule struct{}
type Rules []Rule

type Set map[string]struct{}

func ParseRules(content string, fields, operations Set) Rules {
	parser := newRulesParser(content, fields, operations)
	return parser.parse()
}

type rulesParser struct {
	content    string
	fields     Set
	operations Set
	tokens     *Tokens
}

func newRulesParser(content string, fields, operations Set) *rulesParser {
	return &rulesParser{
		content:    content,
		fields:     fields,
		operations: operations,
		tokens:     &Tokens{},
	}
}

func (p *rulesParser) parse() Rules {
	return Rules{}
}
