package dependancy

import (
	"golang.org/x/exp/maps"
)

type wrapper struct {
	id         string
	impl       Node
	precursors []*wrapper
}

type Graph struct {
	all map[string]*wrapper
}

func NewGraph() *Graph {
	return &Graph{
		all: make(map[string]*wrapper),
	}
}

func (f *Graph) AddNode(nodes ...Node) {
	for _, node := range nodes {
		f.all[node.ID()] = &wrapper{
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

func (f *Graph) SetPrecursorNodes(node Node, precursors ...Node) error {
	nodeId := node.ID()
	precuserIds := idsFromNodes(precursors...)
	return f.SetPrecursorIDs(nodeId, precuserIds...)
}

func (f *Graph) SetPrecursorIDs(id string, precursors ...string) error {
	parent, ok := f.all[id]
	if !ok {
		return newError("unknow").withNodeIDs(id)
	}
	parent.precursors = nil
	for _, precursor := range precursors {
		child, ok := f.all[precursor]
		if !ok {
			return newError("unknown").withNodeIDs(precursor)
		}
		parent.precursors = append(parent.precursors, child)
	}
	return nil
}

func (f *Graph) PlanNodes(targets ...Node) ([]Node, error) {
	targetIds := idsFromNodes(targets...)
	return f.PlanIDs(targetIds...)
}

func (f *Graph) PlanIDs(targetIds ...string) (NodeArray, error) {
	plan := make(NodeArray, len(targetIds))
	todoSet := make(map[string]*wrapper, len(targetIds))

	for _, targetId := range targetIds {
		target, ok := f.all[targetId]
		if !ok {
			return nil, newError("unknown").withNodeIDs(targetId)
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
					todoSet[precursor.id] = precursor
				}
			}
			if precursorsAllPanned {
				plan = append(plan, todo.impl)
				delete(todoSet, todo.id)
				progress = true
			}
		}
		if !progress && len(todoSet) > 0 {
			keys := maps.Keys(todoSet)
			return nil, newError("circular dependancy").withNodeIDs(keys...)
		}
	}

	return plan, nil
}
