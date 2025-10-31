package tools

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	Query   string
	Clauses []string
	Values  []any
}

// Add Clause to Query
func (f *QueryBuilder) AddClause(column string, Value any) {
	f.Clauses = append(f.Clauses, fmt.Sprintf("%s = $%d", column, len(f.Values)+1))
	f.Values = append(f.Values, Value)
}

// Build SQL String
func (f *QueryBuilder) Build() string {
	return fmt.Sprintf(f.Query, strings.Join(f.Clauses, ", "))
}
