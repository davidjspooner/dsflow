package job

type Node interface {
	ID() string
}

type NodeArray []Node

func (na NodeArray) ContainsNode(node Node) bool {
	for _, n := range na {
		if n == node {
			return true
		}
	}
	return false
}

func (na NodeArray) ContainsID(id string) bool {
	for _, n := range na {
		if n.ID() == id {
			return true
		}
	}
	return false
}
