package database

import (
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

type Schema struct {
	Tables    map[string]*types.Table
	Relations []types.Relationship
}

type Table struct {
	Name    string
	Columns []Column
	Count   int
}

type Column struct {
	Name      string
	Type      string
	Generator string
	Params    map[string]interface{}
}

type Relation struct {
	FromTable  string
	FromColumn string
	ToTable    string
	ToColumn   string
}
