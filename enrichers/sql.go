package enrichers

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	// Using go sql driver
)

// SQLClient SQL Client
type SQLClient struct {
	config       *mysql.Config
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
	client.config, err = mysql.ParseDSN(urlString)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Connect to SQL server
func (client *SQLClient) Connect(ctx context.Context) error {
	db, err := sql.Open("mysql", client.url.String())
	if err != nil {
		return err
	}

	// Open doesn't open a connection. Validate DSN data:
	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	client.db = db
	return nil
}

// GetIP Get IP of SQL server
func (client *SQLClient) GetIP() net.IP {
	if colon := strings.Index(client.config.Addr, ":"); colon != -1 {
		return net.ParseIP(client.config.Addr[0:colon])
	} else {
		return net.ParseIP(client.config.Addr)
	}
}

// GetPort Get port of SQL server
func (client *SQLClient) GetPort() uint16 {
	p := ""
	if colon := strings.Index(client.config.Addr, ":"); colon != -1 {
		p = client.config.Addr[colon+1:]
	} else {
		p = "3306"
	}
	// Parse
	port, err := strconv.ParseUint(p, 10, 16)
	if err != nil {
		return 3306
	}
	return uint16(port)
}

// GetConnectString Get connect string
func (client *SQLClient) GetConnectString() string {
	return client.url.String()
}

// IsConnected Is server connected.  Will attempt to open a connection
func (client *SQLClient) IsConnected() bool {
	if client.db == nil {
		return false
	}
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
	// TODO: This will panic if connect fails
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
