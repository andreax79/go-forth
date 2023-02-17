package fcpu

type Statement uint8

const (
	None Statement = iota
	If
	Else
)

type ContextStack struct {
	node *Node
	ids  map[Statement]int
}

type Node struct {
	statement Statement
	id        int
	next      *Node
}

// Enter a new context
func (s *ContextStack) Enter(statement Statement) {
	// Inizialize the id map
	if s.ids == nil {
		s.ids = map[Statement]int{}
	}
	// Get and incmente the id
	id, exists := s.ids[statement]
	if exists {
		id++
	} else {
		id = 1
	}
	s.ids[statement] = id
	// Push the node into the stack
	s.node = &Node{
		statement: statement,
		id:        id,
		next:      s.node,
	}
}

// Check the current context
func (s *ContextStack) Is(statement Statement) bool {
	if s.node != nil {
		return s.node.statement == statement
	}
	return None == statement
}

// Get the id of the current context
func (s *ContextStack) Id() int {
	if s.node != nil {
		return s.node.id
	}
	return 0
}

// Leave the current context
func (s *ContextStack) Exit() {
	if s.node != nil {
		s.node = s.node.next
	}
}

// Change the current context statement
func (s *ContextStack) Change(statement Statement) {
	if s.node != nil {
		s.node.statement = statement
	}
}
