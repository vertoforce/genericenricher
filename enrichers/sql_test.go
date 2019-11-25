package enrichers

import (
	"context"
	"fmt"
	"io"
	"testing"
)

func TestNewSQL(t *testing.T) {
	sql, err := NewSQL("root:pass@tcp(127.0.0.1:3306)/test")
	if err != nil {
		t.Errorf("Error creating sql server")
		return
	}

	tables := sql.GetTables()
	if len(tables) == 0 {
		t.Errorf("Did not get any tables")
	}

	columnNames, rows := sql.GetRows(context.Background(), tables[0])

	// Check column names
	if len(columnNames) == 0 {
		t.Errorf("No column names")
	}

	// Check rows
	totalRows := 0
	for range rows {
		totalRows++
		break
	}
	if totalRows == 0 {
		t.Errorf("Did not get any rows")
	}
}

func TestDump(t *testing.T) {
	sql, err := NewSQL("root:pass@tcp(127.0.0.1:3306)/test")
	if err != nil {
		t.Errorf("Error creating sql server")
		return
	}

	reader, err := sql.Dump(context.Background())
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	p := make([]byte, 1024)
	read, err := reader.Read(p)
	if err != nil && err != io.EOF {
		t.Errorf("Error reading")
		fmt.Printf("Read %d bytes\n", read)
	}
}
