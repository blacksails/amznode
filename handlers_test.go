package amznode_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/blacksails/amznode"
	"github.com/blacksails/amznode/pg"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (http.Handler, func(func(*testing.T)) func(*testing.T)) {
	dbUser := amznode.GetEnv("POSTGRES_USER", "postgres")
	dbPass := amznode.GetEnv("POSTGRES_PASS", "postgres")
	dbName := amznode.GetEnv("POSTGRES_DB", "postgres")
	dbSchema := amznode.GetEnv("POSTGRES_SCHEMA", "amznode")
	dbHost := amznode.GetEnv("POSTGRES_HOST", "localhost")
	dbPort := amznode.GetEnv("POSTGRES_PORT", "5432")

	dbConnStr := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=%s port=%s sslmode=disable",
		dbUser, dbPass, dbHost, dbName, dbPort)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		t.Fatal(err)
	}

	storage, err := pg.New(dbConnStr)
	if err != nil {
		t.Fatal(err)
	}
	server := amznode.New(storage)

	withReset := func(test func(t *testing.T)) func(t *testing.T) {
		schema := pq.QuoteIdentifier(dbSchema)
		table := fmt.Sprintf("%s.%s", schema, pq.QuoteIdentifier(pg.TableName))
		smts := []string{
			fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, schema),
			fmt.Sprintf(`CREATE SCHEMA %s`, schema),
			fmt.Sprintf(`
				CREATE TABLE %s (
					id SERIAL PRIMARY KEY,
					parentID INTEGER REFERENCES %s (id) NULL,
					name TEXT NOT NULL,
					UNIQUE (parentID, name)
				);`, table, table,
			),
		}
		return func(t *testing.T) {
			for _, smt := range smts {
				_, err := db.Query(smt)
				if err != nil {
					t.Fatal(err)
				}
			}
			test(t)
		}
	}

	return server.Handler(), withReset
}

func TestCreate(t *testing.T) {
	h, withReset := setup(t)
	t.Run("byID", withReset(testCreateByID(h)))
}

func testCreateByID(h http.Handler) func(*testing.T) {
	tests := []struct {
		parentID   int
		name       string
		result     interface{}
		resultCode int
	}{
		{
			parentID: 0, name: "root",
			result: amznode.Node{
				ID:     1,
				Name:   "root",
				RootID: 1,
			},
			resultCode: http.StatusCreated,
		},
		{
			parentID: 1, name: "c1",
			result: amznode.Node{
				ID:       2,
				ParentID: 1,
				Name:     "c1",
				RootID:   1,
				Height:   1,
			},
			resultCode: http.StatusCreated,
		},
		{
			parentID: 3, name: "c1",
			result: amznode.ErrorResponse{
				Error: "Could not find node with ID 3",
			},
			resultCode: http.StatusNotFound,
		},
		{
			parentID: 1, name: "c1",
			result: amznode.ErrorResponse{
				Error: "the name 'c1' has already been taken under the parent with id #1",
			},
			resultCode: http.StatusBadRequest,
		},
		{
			parentID: 1, name: "c2",
			result: amznode.Node{
				ID:       4,
				ParentID: 1,
				Name:     "c2",
				RootID:   1,
				Height:   1,
			},
			resultCode: http.StatusCreated,
		},
		{
			parentID: 4, name: "c1",
			result: amznode.Node{
				ID:       5,
				ParentID: 4,
				Name:     "c1",
				RootID:   1,
				Height:   2,
			},
			resultCode: http.StatusCreated,
		},
		{
			parentID: 4, name: "c2",
			result: amznode.Node{
				ID:       6,
				ParentID: 4,
				Name:     "c2",
				RootID:   1,
				Height:   2,
			},
			resultCode: http.StatusCreated,
		},
	}

	return func(t *testing.T) {
		for i, test := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				r := sendRequest(t, h, "POST", fmt.Sprintf("/%d/%s", test.parentID, test.name))
				assertResponse(t, r, test.resultCode, test.result)
			})
		}
	}
}

func TestGet(t *testing.T) {
	h, withReset := setup(t)
	t.Run("roots", withReset(testGetRoots(h)))
	t.Run("byID", withReset(withTestNodes(testGetByID(h), h)))
}

func testGetRoots(h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		// check that we first get an empty list
		r := sendRequest(t, h, "GET", "/")
		assertListResponse(t, r, http.StatusOK, []amznode.Node{})

		tests := []struct {
			name          string
			expectedRoots []amznode.Node
		}{
			{
				name: "myRoot",
				expectedRoots: []amznode.Node{
					amznode.Node{
						ID:     1,
						Name:   "myRoot",
						RootID: 1,
					},
				},
			},
			{
				name: "myOtherRoot",
				expectedRoots: []amznode.Node{
					amznode.Node{
						ID:     2,
						Name:   "myOtherRoot",
						RootID: 2,
					},
					amznode.Node{
						ID:     1,
						Name:   "myRoot",
						RootID: 1,
					},
				},
			},
		}

		for i, test := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				r = sendRequest(t, h, "POST", fmt.Sprintf("/%s", test.name))
				assert.Equal(t, http.StatusCreated, r.StatusCode)
				r = sendRequest(t, h, "GET", "/")
				assertListResponse(t, r, http.StatusOK, test.expectedRoots)
			})
		}
	}
}

func testGetByID(h http.Handler) func(*testing.T) {
	tests := []struct {
		id         int
		resultCode int
		result     interface{}
	}{
		{
			id:         42,
			resultCode: http.StatusNotFound,
			result: amznode.ErrorResponse{
				Error: "Could not find node with ID 42",
			},
		},
		{
			id:         -1,
			resultCode: http.StatusBadRequest,
			result: amznode.ErrorResponse{
				Error: "ids must be greater than or equal 0",
			},
		},
		{
			id:         1,
			resultCode: http.StatusOK,
			result: amznode.Node{
				ID:       1,
				ParentID: 0,
				Name:     "root",
				RootID:   1,
				Height:   0,
				Children: []*amznode.Node{
					{
						ID:       2,
						ParentID: 1,
						Name:     "c1",
						RootID:   1,
						Height:   1,
					},
					{
						ID:       3,
						ParentID: 1,
						Name:     "c2",
						RootID:   1,
						Height:   1,
					},
				},
			},
		},
		{
			id:         6,
			resultCode: http.StatusOK,
			result: amznode.Node{
				ID:       6,
				ParentID: 5,
				Name:     "c5",
				RootID:   1,
				Height:   3,
				Children: []*amznode.Node{
					{
						ID:       7,
						ParentID: 6,
						Name:     "c6",
						RootID:   1,
						Height:   4,
					},
				},
			},
		},
		{
			id:         7,
			resultCode: http.StatusOK,
			result: amznode.Node{
				ID:       7,
				ParentID: 6,
				Name:     "c6",
				RootID:   1,
				Height:   4,
			},
		},
	}

	return func(t *testing.T) {
		for i, test := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				r := sendRequest(t, h, "GET", fmt.Sprintf("/%d", test.id))
				assertResponse(t, r, test.resultCode, test.result)
			})
		}
	}
}

func TestChangeParent(t *testing.T) {
	h, withReset := setup(t)

	tests := []struct {
		id           int
		newParentID  int
		resultCode   int
		result       interface{}
		resultParent interface{}
	}{
		{
			id: 1, newParentID: 42,
			resultCode: http.StatusNotFound,
			result: amznode.ErrorResponse{
				Error: "Could not find node with ID 42",
			},
		},
		{
			id: 42, newParentID: 1,
			resultCode: http.StatusNotFound,
			result: amznode.ErrorResponse{
				Error: "Could not find node with ID 42",
			},
		},
		{
			id: 1, newParentID: 2,
			resultCode: http.StatusBadRequest,
			result: amznode.ErrorResponse{
				Error: "the node with id 2 is a decendant of the node with id 1",
			},
		},
		{
			id: 6, newParentID: 1,
			resultCode: http.StatusOK,
			resultParent: amznode.Node{
				ID:       1,
				ParentID: 0,
				Name:     "root",
				RootID:   1,
				Height:   0,
				Children: []*amznode.Node{
					&amznode.Node{
						ID:       2,
						ParentID: 1,
						Name:     "c1",
						RootID:   1,
						Height:   1,
					},
					&amznode.Node{
						ID:       3,
						ParentID: 1,
						Name:     "c2",
						RootID:   1,
						Height:   1,
					},
					&amznode.Node{
						ID:       6,
						ParentID: 1,
						Name:     "c5",
						RootID:   1,
						Height:   1,
					},
				},
			},
		},
	}

	testFunc := func(t *testing.T) {
		for i, test := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				r := sendRequest(t, h, "PUT", fmt.Sprintf("/%d?parentID=%d", test.id, test.newParentID))

				assertResponse(t, r, test.resultCode, test.result)

				if test.resultParent == nil {
					return
				}

				r = sendRequest(t, h, "GET", fmt.Sprintf("/%d", test.newParentID))
				assertResponse(t, r, http.StatusOK, test.resultParent)
			})
		}
	}

	withReset(withTestNodes(testFunc, h))(t)
}

func TestDelete(t *testing.T) {
	h, withReset := setup(t)

	tests := []struct {
		id         int
		resultCode int
		deletedIDs []int
	}{
		{id: -1, resultCode: http.StatusBadRequest},
		{id: 7, resultCode: http.StatusOK, deletedIDs: []int{7}},
		{id: 5, resultCode: http.StatusOK, deletedIDs: []int{5, 6}},
		{id: 1, resultCode: http.StatusOK, deletedIDs: []int{1, 2, 3, 4}},
	}

	testFunc := func(t *testing.T) {
		for i, test := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				r := sendRequest(t, h, "DELETE", fmt.Sprintf("/%d", test.id))

				assertResponse(t, r, test.resultCode, nil)

				for _, id := range test.deletedIDs {
					r = sendRequest(t, h, "GET", fmt.Sprintf("/%d", id))
					if r.StatusCode != http.StatusNotFound {
						t.Errorf("expected node with id to be not found, got code: %d", r.StatusCode)
					}
				}
			})
		}
	}

	withReset(withTestNodes(testFunc, h))(t)
}

func createTestNodes(t *testing.T, h http.Handler) {
	// 1(root)
	//   2(c1)
	//     4(c3)
	//     5(c4)
	//       6(c5)
	//         7(c6)
	//   3(c2)
	nodes := []struct {
		parentID int
		name     string
	}{
		{parentID: 0, name: "root"}, // id: 1
		{parentID: 1, name: "c1"},   // id: 2
		{parentID: 1, name: "c2"},   // id: 3
		{parentID: 2, name: "c3"},   // id: 4
		{parentID: 2, name: "c4"},   // id: 5
		{parentID: 5, name: "c5"},   // id: 6
		{parentID: 6, name: "c6"},   // id: 7
	}

	// create nodes
	for i, node := range nodes {
		r := sendRequest(t, h, "POST", fmt.Sprintf("/%d/%s", node.parentID, node.name))
		assert.Equal(t, http.StatusCreated, r.StatusCode, i)
	}
}

func sendRequest(t *testing.T, handler http.Handler, method, path string) *http.Response {
	w := httptest.NewRecorder()
	r, err := http.NewRequest(method, path, nil)
	assert.NoError(t, err, "could not create http request")
	handler.ServeHTTP(w, r)
	return w.Result()
}

func assertResponse(t *testing.T, r *http.Response, expectedCode int, expectedBody interface{}) {
	assertStatusCode(t, r, expectedCode)
	if expectedBody == nil {
		return
	}

	if isErrorCode(expectedCode) {
		assertErrorResponse(t, r, expectedBody)
		return
	}

	var respBody amznode.Node
	err := json.NewDecoder(r.Body).Decode(&respBody)
	assert.NoError(t, err, "could not decode json")
	assert.Equal(t, expectedBody, respBody)
}

func assertListResponse(t *testing.T, r *http.Response, expectedCode int, expectedBody interface{}) {
	assertStatusCode(t, r, expectedCode)

	if expectedBody == nil {
		return
	}

	if isErrorCode(expectedCode) {
		assertErrorResponse(t, r, expectedBody)
		return
	}

	var respBody []amznode.Node
	err := json.NewDecoder(r.Body).Decode(&respBody)
	assert.NoError(t, err, "could not decode json")
	assert.Equal(t, expectedBody, respBody)
}

func assertErrorResponse(t *testing.T, r *http.Response, expectedBody interface{}) {
	var respBody amznode.ErrorResponse
	err := json.NewDecoder(r.Body).Decode(&respBody)
	assert.NoError(t, err, "could not decode json")
	assert.Equal(t, expectedBody, respBody)
	return
}

func assertStatusCode(t *testing.T, r *http.Response, expected int) {
	assert.Equal(t, expected, r.StatusCode)
}

func isErrorCode(code int) bool {
	return code < 200 || code >= 300
}

func withTestNodes(test func(*testing.T), h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		createTestNodes(t, h)
		test(t)
	}
}
