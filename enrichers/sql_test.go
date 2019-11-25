package enrichers

import "testing"

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

	rows := sql.GetRows(tables[0])
	if len(rows) == 0 {
		t.Errorf("Did not get any rows")
	}
}
