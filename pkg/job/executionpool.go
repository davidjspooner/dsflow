package job

import (
	"context"
	"fmt"
)

type ExecutionFunction func(ctx context.Context, node Node) error

type Executor interface {
	Run(ctx context.Context, nodes ...Node) ErrorList //Error list implements error itself
}

type executorImpl struct {
	maxParralel int
	fn          ExecutionFunction
}

var _ Executor = &executorImpl{}

func NewExecutor(maxParralel int, f ExecutionFunction) Executor {
	return &executorImpl{
		maxParralel: maxParralel,
		fn:          f,
	}
}

func (ep *executorImpl) safeFnCall(ctx context.Context, node Node) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
		if err != nil {
			wrappedErr := newError().withNodeIDs(node.ID()).withCause(err)
			if r != nil {
				wrappedErr = wrappedErr.withStack()
			}
			err = wrappedErr
		}
	}()
	return ep.fn(ctx, node)
}

func (ep *executorImpl) worker(ctx context.Context, todo <-chan Node, done chan<- error) {
	for node := range todo {
		err := ep.safeFnCall(ctx, node)
		if err != nil {
			done <- err
			return
		}
	}
	done <- nil
}

func (ep *executorImpl) Run(ctx context.Context, nodes ...Node) ErrorList {
	todo := make(chan Node, len(nodes))
	done := make(chan error, len(nodes))
	for _, node := range nodes {
		todo <- node
	}
	close(todo)

	for i := 0; i < ep.maxParralel; i++ {
		go ep.worker(ctx, todo, done)
	}

	errorList := ErrorList{}
	for i := 0; i < len(nodes); i++ {
		err := <-done
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if len(errorList) > 0 {
		return errorList
	}
	return nil
}
