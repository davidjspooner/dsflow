package job

import (
	"golang.org/x/exp/maps"
)

type nodeWrapper struct {
	id         string
	impl       Node
	precursors []*nodeWrapper
}

type NodeGraph struct {
	all map[string]*nodeWrapper
}

func NewNodeGraph() *NodeGraph {
	return &NodeGraph{
		all: make(map[string]*nodeWrapper),
	}
}

func (f *NodeGraph) AddNode(nodes ...Node) {
	for _, node := range nodes {
		f.all[node.ID()] = &nodeWrapper{
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

func (f *NodeGraph) SetPrecursorNodes(node Node, precursors ...Node) error {
	nodeId := node.ID()
	precuserIds := idsFromNodes(precursors...)
	return f.SetPrecursorIDs(nodeId, precuserIds...)
}

func (f *NodeGraph) SetPrecursorIDs(id string, precursors ...string) error {
	parent, ok := f.all[id]
	if !ok {
		return newError().withNodeIDs(id).withCausef("unknown node")
	}
	parent.precursors = nil
	for _, precursor := range precursors {
		child, ok := f.all[precursor]
		if !ok {
			return newError().withNodeIDs(precursor).withCausef("unknown precursor")
		}
		parent.precursors = append(parent.precursors, child)
	}
	return nil
}

func (f *NodeGraph) PlanNodes(targets ...Node) (NodeArray, error) {
	targetIds := idsFromNodes(targets...)
	return f.PlanIDs(targetIds...)
}

func (f *NodeGraph) PlanIDs(targetIds ...string) (NodeArray, error) {
	plan := make(NodeArray, len(targetIds))
	todoSet := make(map[string]*nodeWrapper, len(targetIds))

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
			return nil, newError().withNodeIDs(keys...).withCausef("circular dependency")
		}
	}

	return plan, nil
}
