package amznode

// Storage is our main storage interface
type Storage interface {
	// Create creates a new node with the given `name` and `parentID`. If
	// `parentID` is set to `0`, then the node will be a root node.
	//
	// If the parent could not be found an `ErrNotFound` error will be
	// returned. If already exists a node with the given `name` and `parentID`
	// an `ErrNameTaken` error will be returned.
	Create(name string, parentID int) (*Node, error)

	// Get gets the node with the given `id` along with its children.
	//
	// If the Node could not be found an `ErrNotFound` will be returned.
	Get(id int) (*Node, error)

	// GetRoots get all the tree roots
	GetRoots() ([]*Node, error)

	// ChangeParent changes the parent of the node with `id` to the node with
	// the `newParentID`.
	//
	// If either of the nodes does not exist an `ErrNotFound` error will be
	// returned. If the node with id `newParentID` already has a node with the
	// name of the node with id of `id`. Then an `ErrNameTaken` will be
	// returned.
	ChangeParent(id, newParentID int) error

	// Delete deletes a node along with all its decendent children.
	//
	// If a node with id of `id` could not be found then an `ErrNotFound` will
	// be returned.
	Delete(id int) error

	//CreatePath(path string) (*Node, error)
	//Get(path string) (*Node, error)
	//DeleteByPath(path string) error
	//Delete(node *Node) error
	//GetByIDRec(id int) (*Node, error)
	//GetByPathRec(path string) (*Node, error)
}
