package dependancy

import (
	"fmt"
	"strings"
)

type Error struct {
	message string
	nodes   []string
}

func (e *Error) Error() string {
	nodeString := strings.Join(e.nodes, ", ")
	return e.message + ": " + nodeString
}

func newError(format string, args ...any) *Error {
	return &Error{
		message: fmt.Sprintf(format, args...),
	}
}

func (e *Error) withNodeIDs(ids ...string) *Error {
	e.nodes = append(e.nodes, ids...)
	return e
}
