package amznode

import (
	"fmt"
	"net/http"
)

// ErrNotFound is returned when a node could not be found in the storage
type ErrNotFound struct {
	ID int
}

// NewErrNotFound returns a new ErrNotFound
func NewErrNotFound(id int) *ErrNotFound {
	return &ErrNotFound{ID: id}
}

func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("Could not find node with ID %d", err.ID)
}

// ErrNameTaken is returned when a chil node has a sibling with a conflicting
// name.
type ErrNameTaken struct {
	Name     string
	ParentID int
}

// NewErrNameTaken instantiates a ErrNameTaken error
func NewErrNameTaken(name string, parentID int) *ErrNameTaken {
	return &ErrNameTaken{Name: name, ParentID: parentID}
}

func (err *ErrNameTaken) Error() string {
	return fmt.Sprintf("the name '%s' has already been taken under the parent with id #%d", err.Name, err.ParentID)
}

// NewErrNodeIsDecendant instantiates a ErrNodeIsDecendant error
func NewErrNodeIsDecendant(id, decendantID int) *ErrNodeIsDecendant {
	return &ErrNodeIsDecendant{ID: id, DecendantID: decendantID}
}

// ErrNodeIsDecendant is returned when an update of parent fails because the
// new parent is a decendant of the node which is updated.
type ErrNodeIsDecendant struct {
	ID          int
	DecendantID int
}

func (err *ErrNodeIsDecendant) Error() string {
	return fmt.Sprintf(
		"the node with id %d is a decendant of the node with id %d",
		err.DecendantID, err.ID,
	)
}

func handleStorageError(w http.ResponseWriter, r *http.Request, err error) {
	switch err.(type) {
	case *ErrNotFound:
		respondErr(w, r, err, http.StatusNotFound)
		return
	case *ErrNameTaken:
		respondErr(w, r, err, http.StatusBadRequest)
		return
	case *ErrNodeIsDecendant:
		respondErr(w, r, err, http.StatusBadRequest)
		return
	default:
		respondErr(w, r, err, http.StatusInternalServerError)
		return
	}
}
