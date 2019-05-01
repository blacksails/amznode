package pg

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/blacksails/amznode"
)

// TableName is the name of the node table in the database
const TableName = "nodes"
const nodeCols = "id, parentID, rootID, name, height"

// Storage is an implementaion of the `amznode.Storage` interface backed by
// PostgreSQL
type Storage struct {
	db     *sql.DB
	schema string
}

func (s Storage) table() string {
	return fmt.Sprintf(
		"%s.%s", pq.QuoteIdentifier(s.schema), pq.QuoteIdentifier(TableName))
}

// New instantiates a new Storage based on the given dataSourceName
// string
func New(dataSourceName string) (*Storage, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	// TODO: ensure that db and table is created
	return &Storage{db: db, schema: "amznode"}, nil
}

// NewFromEnv instantiates a new pg.Storage based on the following env
// variables.
//
// - POSTGRES_USER
// - POSTGRES_PASS
// - POSTGRES_DB
// - POSTGRES_HOST
// - POSTGRES_PORT
func NewFromEnv() (*Storage, error) {
	dbUser := amznode.GetEnv("POSTGRES_USER", "postgres")
	dbPass := amznode.GetEnv("POSTGRES_PASS", "postgres")
	dbName := amznode.GetEnv("POSTGRES_DB", "postgres")
	dbSchema := amznode.GetEnv("POSTGRES_SCHEMA", "amznode")
	dbHost := amznode.GetEnv("POSTGRES_HOST", "localhost")
	dbPort := amznode.GetEnv("POSTGRES_PORT", "5432")
	dbConnStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbUser, dbPass, dbName, dbHost, dbPort)
	storage, err := New(dbConnStr)
	if err != nil {
		return storage, err
	}
	storage.schema = dbSchema
	err = ensureSchemaAndTableExists(storage.db, pq.QuoteIdentifier(dbSchema), storage.table())
	return storage, err
}

func ensureSchemaAndTableExists(db *sql.DB, schema, table string) error {
	qs := []string{
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema),
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id SERIAL PRIMARY KEY,
				parentID INTEGER REFERENCES %s (id) NULL,
				name TEXT NOT NULL,
				UNIQUE (parentID, name)
			);`, table, table,
		),
	}
	for _, q := range qs {
		_, err := db.Query(q)
		if err != nil {
			return err
		}
	}
	return nil
}
