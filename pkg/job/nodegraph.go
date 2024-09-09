package job

import (
	"context"

	"golang.org/x/exp/maps"
)

type NodeWithPrecursors struct {
	id         string
	impl       Node
	precursors []*NodeWithPrecursors
}

func (n *NodeWithPrecursors) ID() string {
	return n.id
}

func (n *NodeWithPrecursors) Node() Node {
	return n.impl
}
func (n *NodeWithPrecursors) Precursors() []*NodeWithPrecursors {
	var precursors []*NodeWithPrecursors
	precursors = append(precursors, n.precursors...)
	return precursors
}

type NodeDependancyOrdering []*NodeWithPrecursors

func (ordering NodeDependancyOrdering) ContainsNode(node Node) bool {
	for _, wrapper := range ordering {
		if wrapper.impl == node {
			return true
		}
	}
	return false
}

type NodeGraph interface {
	AddNode(nodes ...Node)
	SetPrecursorNodes(node Node, precursors ...Node) error
	SetPrecursorIDs(id string, precursors ...string) error
	PlanNodes(targets ...Node) (NodeDependancyOrdering, error)
	PlanIDs(targetIds ...string) (NodeDependancyOrdering, error)
}

type nodeGraph struct {
	all map[string]*NodeWithPrecursors
}

func NewNodeGraph() NodeGraph {
	return &nodeGraph{
		all: make(map[string]*NodeWithPrecursors),
	}
}

func (f *nodeGraph) AddNode(nodes ...Node) {
	for _, node := range nodes {
		f.all[node.ID()] = &NodeWithPrecursors{
			id:   node.ID(),
			impl: node,
		}
	}
}

func idsFromNodes(nodes ...Node) []string {
	ids := make([]string, len(nodes))
	for i, node := range nodes {
		ids[i] = node.ID()
	}
	return ids
}

func (f *nodeGraph) SetPrecursorNodes(node Node, precursors ...Node) error {
	nodeId := node.ID()
	precuserIds := idsFromNodes(precursors...)
	return f.SetPrecursorIDs(nodeId, precuserIds...)
}

func (f *nodeGraph) SetPrecursorIDs(id string, precursors ...string) error {
	parent, ok := f.all[id]
	if !ok {
		return newError().withNodeIDs(id).withCausef("unknown")
	}
	parent.precursors = nil
	for _, precursor := range precursors {
		child, ok := f.all[precursor]
		if !ok {
			return newError().withNodeIDs(precursor).withCausef("unknown")
		}
		parent.precursors = append(parent.precursors, child)
	}
	return nil
}

func (f *nodeGraph) PlanNodes(targets ...Node) (NodeDependancyOrdering, error) {
	targetIds := idsFromNodes(targets...)
	return f.PlanIDs(targetIds...)
}

func (f *nodeGraph) PlanIDs(targetIds ...string) (NodeDependancyOrdering, error) {
	plan := make(NodeDependancyOrdering, 0, len(targetIds))
	todoSet := make(map[string]*NodeWithPrecursors, len(targetIds))

	for _, targetId := range targetIds {
		target, ok := f.all[targetId]
		if !ok {
			return nil, newError().withNodeIDs(targetId).withCausef("unknown target")
		}
		todoSet[targetId] = target
	}

	for len(todoSet) > 0 {
		progress := false
		for _, todo := range todoSet {
			precursorsAllPanned := true
			for _, precursor := range todo.precursors {
				if !plan.ContainsNode(precursor.impl) {
					precursorsAllPanned = false
					if _, ok := todoSet[precursor.id]; !ok {
						todoSet[precursor.id] = precursor
						progress = true
					}
				}
			}
			if precursorsAllPanned {
				plan = append(plan, todo)
				delete(todoSet, todo.id)
				progress = true
			}
		}
		if !progress && len(todoSet) > 0 {
			keys := maps.Keys(todoSet)
			return nil, newError().withNodeIDs(keys...).withCausef("circular dependency")
		}
	}

	return plan, nil
}

func (ordering *NodeDependancyOrdering) Run(ctx context.Context, maxParralel int, fn func(ctx context.Context, node *NodeWithPrecursors) error, logger Logger) error {

	executer := NewExecuter[*NodeWithPrecursors](logger)

	done := make(map[string]chan bool)
	for _, node := range *ordering {
		done[node.ID()] = make(chan bool)
	}

	executer.Start(ctx, maxParralel, func(ctx context.Context, nwp *NodeWithPrecursors) error {
		defer close(done[nwp.ID()])
		for {
			allDependanciesDone := true
			for _, dependancy := range nwp.precursors {
				select {
				case <-done[dependancy.ID()]:
					err := executer.GetStatus(dependancy)
					if err != nil {
						if err == errPending {
							allDependanciesDone = false
						} else {
							return newError().withNodeIDs(nwp.ID()).withCausef("dependancy %q failed", dependancy.ID())
						}
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if allDependanciesDone {
				return fn(ctx, nwp)
			}
		}
	}, *ordering)
	err := executer.WaitForCompletion()
	return err
}
