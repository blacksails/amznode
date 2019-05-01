package pg

import (
	"database/sql"

	"github.com/blacksails/amznode"
)

type node struct {
	id       int
	parentID sql.NullInt64
	name     string
}

func (n node) ToDomain() *amznode.Node {
	return &amznode.Node{
		ID:       n.id,
		ParentID: int(n.parentID.Int64),
		Name:     n.name,
	}
}
