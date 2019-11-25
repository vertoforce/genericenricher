package enrichers

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// SQLClient SQL Client
type SQLClient struct {
	db *sql.DB
}

// NewSQL Create new SQL client
func NewSQL(urlString string) (*SQLClient, error) {
	db, err := sql.Open("mysql", urlString)
	if err != nil {
		return nil, err
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	client := &SQLClient{db: db}

	return client, nil
}

// Close the connection
func (client *SQLClient) Close() {
	client.db.Close()
}

// GetTables Get mysql table names
func (client *SQLClient) GetTables() []string {
	rows, err := client.db.Query("SHOW TABLES")
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	tables := []string{}
	var tableName string
	for rows.Next() {
		rows.Scan(&tableName)
		tables = append(tables, tableName)
	}

	return tables
}

// GetRows Get rows of data in table.  NOTE: tableName is NOT sanitized, it is injected right in to the query
func (client *SQLClient) GetRows(tableName string) []string {
	rows, err := client.db.Query("SELECT * FROM " + tableName)
	if err != nil {
		return nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil
	}

	row := make([]interface{}, len(columns))
	for i := range columns {
		row[i] = new(string)
	}

	for rows.Next() {
		rows.Scan(row...)
	}

	return []string{}
}
