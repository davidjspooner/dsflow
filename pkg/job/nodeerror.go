package job

import (
	"fmt"
	"runtime/debug"
	"strings"
)

type NodeIdError struct {
	Stack   string
	Cause   error
	NodeIds []string
}

func (e *NodeIdError) Error() string {
	sb := strings.Builder{}
	for n, id := range e.NodeIds {
		if n > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%q", id)
	}
	return sb.String() + ": " + e.Cause.Error()
}

func newError() *NodeIdError {
	return &NodeIdError{}
}

func (e *NodeIdError) withNodeIDs(ids ...string) *NodeIdError {
	e.NodeIds = append(e.NodeIds, ids...)
	return e
}

func (e *NodeIdError) withStack() *NodeIdError {
	e.Stack = string(debug.Stack())
	return e
}

func (e *NodeIdError) withCause(err error) *NodeIdError {
	e.Cause = err
	return e
}

func (e *NodeIdError) withCausef(format string, args ...interface{}) *NodeIdError {
	return e.withCause(fmt.Errorf(format, args...))
}
