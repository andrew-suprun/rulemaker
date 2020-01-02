package model

import (
	"league.com/rulemaker/msg"
	"league.com/rulemaker/util"
)

type Connector interface {
	FetchFiles() (result Files, err error)
}

type RuleEngine interface {
	IngestFile(fileName string, records Records, mappingRulesPath, mergingRulesPath string, config msg.M) error
	Entries() Entries
}

type Processor interface {
	Process(entry *Entry)
	Done()
}

type File struct {
	Name    string
	Type    string
	Content []byte
}

type Files []File

type Record interface {
	Line() int
	Field(name, kind string) (interface{}, error)
}

type Records []Record

type Entity msg.M

func (e Entity) EntityId() string {
	id, _ := e["employee_id"].(string)
	return id
}

type Action string

var (
	None   Action = "none"
	Log    Action = "log"
	Ticket Action = "ticket"
	Fail   Action = "fail"
	Alert  Action = "alert"
	Skip   Action = "skip"
)

type Diagnostic struct {
	Message string
	Field   string
	Action  Action
}

type Diagnostics []Diagnostic

type Source struct {
	FilePath   string `bson:"file_path"`
	LineNumber int    `bson:"line_number"`
}

type Sources []Source

type Entries map[string]*Entry
type Entry struct {
	Entity      Entity
	Sources     Sources
	Diagnostics Diagnostics
}

func (e *Entry) String() string {
	return util.ToJson(e)
}

func (e *Entry) Report(message string, field string, action Action) {
	e.Diagnostics = append(e.Diagnostics, Diagnostic{Message: message, Field: field, Action: action})
}
