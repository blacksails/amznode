package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/blacksails/amznode"
	"github.com/lib/pq"
)

const codeUniqueViolation = "23505"

// Create implements `amznode.Storage.Create`
func (s *Storage) Create(name string, parentID int) (*amznode.Node, error) {

	n := node{name: name}

	if parentID != 0 {
		// when we have a child node we need to fetch the parent to see
		// that it exists and get the id of the root node.
		parent, err := s.Get(parentID)
		if err != nil {
			return nil, err
		}
		n.parentID = sql.NullInt64{Int64: int64(parent.ID), Valid: true}
	}

	q := fmt.Sprintf(`
		INSERT INTO %s (parentID, name)
		VALUES ($1, $2) RETURNING id`,
		s.table(),
	)

	err := s.db.QueryRow(q, n.parentID, name).Scan(&n.id)
	if err, ok := err.(*pq.Error); ok {
		switch err.Code {
		case codeUniqueViolation:
			return nil, amznode.NewErrNameTaken(name, parentID)
		}
	}
	if err != nil {
		return nil, err
	}

	return s.Get(n.id)
}

// Get implements `amznode.Storage.Get`
func (s *Storage) Get(id int) (*amznode.Node, error) {
	q := fmt.Sprintf(`
		WITH RECURSIVE q AS (
			SELECT h.*, 1 AS level
			FROM %s h
			WHERE id = $1 OR parentID = $1
			UNION ALL
			SELECT hp.*, level + 1
			FROM q
			JOIN %s hp
			ON hp.id = q.parentID
		)
		SELECT id, parentID, name
		FROM q
		ORDER BY level DESC
	`, s.table(), s.table())

	rows, err := s.db.Query(q, id)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	_, nodesByID, err := loadRawNodes(rows)
	node, ok := nodesByID[id]
	if !ok {
		return nil, amznode.NewErrNotFound(id)
	}
	return node, nil
}

// GetRoots implements `amznode.Storage.GetRoots`
func (s *Storage) GetRoots() ([]*amznode.Node, error) {
	t := s.table()
	q := fmt.Sprintf(`
		WITH q AS (
			SELECT * 
			FROM %s
			WHERE parentID IS NULL
		)
		SELECT * FROM q
		UNION ALL
		SELECT * FROM %s WHERE parentID IN (SELECT id FROM q)
	`, t, t)

	rows, err := s.db.Query(q)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	rootIDs, nodesByID, err := loadRawNodes(rows)
	roots := make([]*amznode.Node, len(rootIDs))
	for i, id := range rootIDs {
		roots[i] = nodesByID[id]
	}
	sort.SliceStable(roots, func(i, j int) bool {
		return roots[i].Name < roots[j].Name
	})

	return roots, nil
}

// ErrInvalidTree is returned when the data in the database couldn't be parsed
// to a valid tree. This should be considered an internal error as this should
// never happen.
var ErrInvalidTree = errors.New("invalid tree")

func loadRawNodes(rows *sql.Rows) ([]int, map[int]*amznode.Node, error) {
	rootIDs := []int{}
	nodesByID := map[int]*amznode.Node{}

	// Load rows from SQL to domain Nodes
	for rows.Next() {
		var n node
		if err := rows.Scan(&n.id, &n.parentID, &n.name); err != nil {
			return nil, nil, err
		}
		an := n.ToDomain()
		nodesByID[n.id] = an
	}

	// Register rootIDs and set children
	for _, node := range nodesByID {
		if node.IsRoot() {
			rootIDs = append(rootIDs, node.ID)
			node.RootID = node.ID
			continue
		}
		parent, ok := nodesByID[node.ParentID]
		if !ok {
			return nil, nil, ErrInvalidTree
		}
		parent.Children = append(parent.Children, node)
	}

	// Recursively build nodes
	for _, rootID := range rootIDs {
		node := nodesByID[rootID]
		if node.IsRoot() {
			buildRootNode(node)
		}
	}

	return rootIDs, nodesByID, nil
}

func buildRootNode(rootNode *amznode.Node) {
	var wg sync.WaitGroup
	var auxFunc func(node *amznode.Node, height int)
	auxFunc = func(node *amznode.Node, height int) {
		node.Height = height
		node.RootID = rootNode.ID
		// this sort is mostly here to ensure deterministic behavior, which
		// makes testing easier.
		sort.SliceStable(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
		for _, child := range node.Children {
			wg.Add(1)
			go func(node *amznode.Node) {
				defer wg.Done()
				auxFunc(node, height+1)
			}(child)
		}
	}
	auxFunc(rootNode, 0)
	wg.Wait()
}

// ChangeParent implements amznode.Storage.ChangeParent
func (s *Storage) ChangeParent(id, newParentID int) error {
	_, err := s.Get(id)
	if err != nil {
		return err
	}
	_, err = s.Get(newParentID)
	if err != nil {
		return err
	}

	isDecendant, err := s.isDecendant(id, newParentID)
	if err != nil {
		return err
	}
	if isDecendant {
		return amznode.NewErrNodeIsDecendant(id, newParentID)
	}

	q := fmt.Sprintf("UPDATE %s SET parentID = $1 WHERE id = $2", s.table())
	_, err = s.db.Exec(q, newParentID, id)
	return err
}

func (s *Storage) isDecendant(treeRootID, id int) (bool, error) {
	q := fmt.Sprintf(`
		WITH RECURSIVE q AS (
			SELECT h.*
			FROM %s h
			WHERE id = $1
			UNION ALL
			SELECT hc.*
			FROM q
			JOIN %s hc
			ON q.id = hc.parentID
		)
		SELECT COUNT(*) FROM q WHERE id = $2
	`, s.table(), s.table())

	var count int
	err := s.db.QueryRow(q, treeRootID, id).Scan(&count)
	if err != nil {
		return false, err
	}

	if count != 1 {
		return false, nil
	}
	return true, nil
}

// Delete implements amznode.Storage.Delete
func (s *Storage) Delete(id int) error {
	q := fmt.Sprintf(`
		WITH RECURSIVE q AS (
			SELECT h.* 
			FROM %s h
			WHERE id = $1
			UNION ALL
			SELECT hc.* 
			FROM q 
			JOIN %s hc
			ON q.id = hc.parentID
		)
		DELETE FROM %s WHERE id IN (SELECT id FROM q)`,
		s.table(), s.table(), s.table(),
	)
	_, err := s.db.Exec(q, id)
	return err
}
