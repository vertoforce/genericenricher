package enrichers

import (
	"context"
	"database/sql"
	"io"
	"net"
	"net/url"

	// Using go sql driver
	_ "github.com/go-sql-driver/mysql"
)

// SQLClient SQL Client
type SQLClient struct {
	url          *url.URL
	db           *sql.DB
	reader       io.ReadCloser
	readerCtx    context.Context
	readerCancel context.CancelFunc
}

// NewSQL Create new SQL client
func NewSQL(urlString string) (*SQLClient, error) {
	client := &SQLClient{}

	// Parse URL
	var err error
	client.url, err = url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	err = client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Connect to SQL server
func (client *SQLClient) Connect() error {
	db, err := sql.Open("mysql", client.url.String())
	if err != nil {
		return err
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		return err
	}

	client.db = db

	return nil
}

// GetIP Get IP of SQL server
func (client *SQLClient) GetIP() net.IP {
	// TODO: Fix this
	return urlToIP(client.url)
}

// GetPort Get port of SQL server
func (client *SQLClient) GetPort() uint16 {
	return urlToPort(client.url)
}

// IsConnected Is server connected.  Will attempt to open a connection
func (client *SQLClient) IsConnected() bool {
	return client.db.Ping() == nil
}

// Type Returns SQL
func (client *SQLClient) Type() ServerType {
	return SQL
}

// Close the connection
func (client *SQLClient) Close() error {
	// Stop current reader
	if client.reader != nil {
		client.readerCancel()
		err := client.reader.Close()
		if err != nil {
			return err
		}
	}

	return client.db.Close()
}

func (client *SQLClient) Read(p []byte) (n int, err error) {
	return client.reader.Read(p)
}

// ResetReader reset reader back to initial state
func (client *SQLClient) ResetReader() error {
	// Stop current reader
	if client.reader != nil {
		client.readerCancel()
		err := client.reader.Close()
		if err != nil {
			return err
		}
	}

	// Start new reader
	client.readerCtx, client.readerCancel = context.WithCancel(context.Background())
	reader, err := client.Dump(client.readerCtx)
	if err != nil {
		return err
	}

	client.reader = reader

	return nil
}

// -- SQL specific functions --

// Dump SQL Dump data of entire database
func (client *SQLClient) Dump(ctx context.Context) (io.ReadCloser, error) {
	dumpReader, dumpWriter := io.Pipe()

	go func() {
		defer dumpWriter.Close()

		// For every table
		for _, table := range client.GetTables() {
			// For every row
			_, rows := client.GetRows(ctx, table)
			for row := range rows {
				// For every column
				for _, col := range row {
					// TODO: Cancel with context
					dumpWriter.Write(col)
				}
			}
		}
	}()

	return dumpReader, nil
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
func (client *SQLClient) GetRows(ctx context.Context, tableName string) (columnNames []string, rowsChan chan [][]byte) {
	rows, err := client.db.Query("SELECT * FROM " + tableName)
	if err != nil {
		return nil, nil
	}

	columnNames, err = rows.Columns()
	if err != nil {
		return nil, nil
	}

	rowsChan = make(chan [][]byte)

	go func() {
		defer close(rowsChan)
		defer rows.Close()

		row := make([]interface{}, len(columnNames))
		for i := range columnNames {
			row[i] = new([]byte)
		}

		for rows.Next() {
			rows.Scan(row...)

			// Convert each column to array of bytes
			rowBytes := make([][]byte, len(row))
			for i := range row {
				// Type assert to *[]byte then dereference into []byte
				rowBytes[i] = *(row[i].(*[]byte))
			}

			select {
			case rowsChan <- rowBytes:
			case <-ctx.Done():
				return
			}

		}
	}()

	return columnNames, rowsChan
}
