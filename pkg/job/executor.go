package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Executer[T comparable] interface {
	Start(ctx context.Context, maxParralel int, fn func(context.Context, T) error, jobs []T)
	GetStatus(job T) error
	WaitForCompletion() ErrorList
}

func NewExecuter[T comparable](logger Logger) Executer[T] {
	executer := executer[T]{
		status:  make(map[T]error),
		results: make(chan executerResult[T]),
		logger:  logger,
	}
	return &executer
}

type executerResult[T comparable] struct {
	job T
	err error
}

type executer[T comparable] struct {
	lock    sync.Mutex
	status  map[T]error
	results chan executerResult[T]
	logger  Logger
}

func (ep *executer[T]) setStatus(job T, err error) {
	ep.lock.Lock()
	defer ep.lock.Unlock()
	ep.status[job] = err
}

func (ep *executer[T]) GetStatus(job T) error {
	ep.lock.Lock()
	defer ep.lock.Unlock()
	err, ok := ep.status[job]
	if !ok {
		return errors.New("not found")
	}
	return err
}

var errPending = errors.New("running")

func (ep *executer[T]) safeFnCall(ctx context.Context, job T, fn func(context.Context, T) error) (err error) {
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
			wrappedErr := newError().withCause(err)
			if r != nil {
				wrappedErr = wrappedErr.withStack()
			}
			err = wrappedErr
		}
	}()
	return fn(ctx, job)
}

func (ep *executer[T]) worker(ctx context.Context, todo <-chan T, fn func(context.Context, T) error) {
	for job := range todo {

		err := ep.safeFnCall(ctx, job, fn)
		if err != nil {
			ep.results <- executerResult[T]{job, err}
			continue
		}
		ep.results <- executerResult[T]{job, nil}
	}
}

func (ep *executer[T]) Start(ctx context.Context, maxParralel int, fn func(context.Context, T) error, jobs []T) {
	todo := make(chan T, len(jobs))
	for _, job := range jobs {
		ep.setStatus(job, errPending)
		todo <- job
	}
	for i := 0; i < Min(maxParralel, len(jobs)); i++ {
		go ep.worker(ctx, todo, fn)
	}
}

func (ep *executer[T]) WaitForCompletion() ErrorList {
	var errorList ErrorList
	for i := 0; i < len(ep.status); i++ {
		result := <-ep.results
		ep.setStatus(result.job, result.err)
		if result.err != nil {
			errorList = append(errorList, result.err)
		}
	}

	if len(errorList) > 0 {
		return errorList
	}
	return nil
}
