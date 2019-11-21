package enrichers

import (
	"context"
	"fmt"
	"regexmachine"
	"testing"
)

func TestNewElK(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// Check IP and Port
	fmt.Println(con.GetIP())
	if con.GetIP().String() != "127.0.0.1" {
		t.Errorf("Incorrect IP")
	}
	if con.GetPort() != 9200 {
		t.Errorf("Incorrect Port")
	}
	if con.Type() != ELK {
		t.Errorf("Incorrect Type")
	}

	// Check closing
	if con.IsConnected() != true {
		t.Errorf("Should be connected")
	}
	con.Close()
	if con.IsConnected() != false {
		t.Errorf("Should be disconnected")
	}
}

func TestGetIndices(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// Get indices
	indices, err := con.GetIndices()
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// Get matched indices with "match-all" rule
	matchedIndices, err := con.GetIndicesMatchingRules(context.Background(), regexmachine.MatchAll, 500) // Limit to 500 docs to avoid long wait times
	if err != nil {
		t.Errorf("Failed to get indices")
	}
	// Print the indices we matched
	for _, relevantIndex := range matchedIndices {
		fmt.Println(relevantIndex)
	}

	// Make sure we matches all indices
	if len(indices) != len(matchedIndices) {
		t.Errorf("Did not match all indices")
	}
}

// Testing index with more than 1 entry
var testingIndex = ".kibana"

func TestGetData(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// Check to make sure the limit works
	ctx := context.Background()
	dataStream := con.GetData(ctx, testingIndex, 1)

	totalHits := 0
	for range dataStream {
		totalHits++
		if totalHits > 1 {
			t.Errorf("Too many hits, limit didn't work")
			break
		}
	}

	// Check cancel works
	ctx, cancel := context.WithCancel(context.Background())
	dataStream = con.GetData(ctx, testingIndex, -1)

	totalHits = 0
	for range dataStream {
		totalHits++
		cancel()

		if totalHits > 1 {
			t.Errorf("Too many hits, cancel didn't work")
			break
		}
	}
	cancel()

	if totalHits == 0 {
		t.Errorf("No data found on index `%s`", testingIndex)
	}
}
