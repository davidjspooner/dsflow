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
	nodeString := strings.Join(e.NodeIds, ", ")
	return e.Cause.Error() + ": " + nodeString
}

func newError() *NodeIdError {
	return &NodeIdError{
	}
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

