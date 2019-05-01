package amznode

// Node represents a node in the organization tree
type Node struct {
	ID       int     `json:"id"`
	ParentID int     `json:"parent_id,omitempty"`
	Name     string  `json:"name"`
	RootID   int     `json:"root_id"`
	Height   int     `json:"height"`
	Children []*Node `json:"children,omitempty"`
}

// IsRoot returns true if the node does not have a parent
func (n Node) IsRoot() bool {
	return n.ParentID == 0
}
