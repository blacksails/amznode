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

	return buildTree(id, rows)
}

// GetRoots implements `amznode.Storage.GetRoots`
func (s *Storage) GetRoots() ([]*amznode.Node, error) {
	t := s.table()
	q := fmt.Sprintf(`
		WITH roots AS (
			SELECT * 
			FROM %s
			WHERE parentID = NULL
		)
		SELECT * FROM roots
		UNION ALL
		SELECT * FROM %s WHERE parentID IN (SELECT id FROM roots)
	`, t, t)

	rows, err := s.db.Query(q)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ErrInvalidTree is returned when the data in the database couldn't be parsed
// to a valid tree. This should be considered an internal error as this should
// never happen.
var ErrInvalidTree = errors.New("invalid tree")

func buildTree(id int, rows *sql.Rows) (*amznode.Node, error) {
	var root node
	if !rows.Next() {
		return nil, amznode.NewErrNotFound(id)
	}
	if err := rows.Scan(&root.id, &root.parentID, &root.name); err != nil {
		return nil, err
	}
	aroot := root.ToDomain()
	aroot.RootID = aroot.ID
	nodesByID := map[int]*amznode.Node{root.id: aroot}
	for rows.Next() {
		var n node
		if err := rows.Scan(&n.id, &n.parentID, &n.name); err != nil {
			return nil, err
		}
		an := n.ToDomain()
		an.RootID = aroot.ID
		nodesByID[n.id] = an
	}

	for _, node := range nodesByID {
		if node.IsRoot() {
			continue
		}
		parent, ok := nodesByID[node.ParentID]
		if !ok {
			return nil, ErrInvalidTree
		}
		parent.Children = append(parent.Children, node)
		// this sort is mostly here to ensure deterministic behavior, which
		// makes testing easier.
		sort.SliceStable(parent.Children, func(i, j int) bool {
			return parent.Children[i].Name < parent.Children[j].Name
		})
	}

	setHeight(nodesByID[root.id], 0)

	return nodesByID[id], nil
}

func setHeight(node *amznode.Node, height int) {
	node.Height = height
	var wg sync.WaitGroup
	for _, child := range node.Children {
		wg.Add(1)
		go func(node *amznode.Node) {
			defer wg.Done()
			setHeight(node, height+1)
		}(child)
	}
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
