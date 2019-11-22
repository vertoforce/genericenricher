package enrichers

import (
	"context"
	"fmt"
	"io"
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
	indices, err := con.GetIndices(context.Background())
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
var sizeOfTestingELK = 162

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

func TestRead(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// See if we can read in pieces
	// First piece
	p := make([]byte, sizeOfTestingELK/2)
	read, err := con.Read(p)
	if err != nil {
		t.Errorf(err.Error())
	}
	if read != sizeOfTestingELK/2 {
		t.Errorf("Did not read fully into buffer")
	}
	// Second piece
	p = make([]byte, sizeOfTestingELK)
	read, err = con.Read(p)
	if err != io.EOF && err != nil {
		t.Errorf(err.Error())
	}
	if read != sizeOfTestingELK-sizeOfTestingELK/2 {
		t.Errorf("Did not read remaining")
	}

	// See if we can reset and read it all
	con.ResetReader()

	// See if we can read in pieces with 1 byte offset
	// First piece
	p = make([]byte, sizeOfTestingELK-1)
	read, err = con.Read(p)
	if err != nil {
		t.Errorf(err.Error())
	}
	if read != sizeOfTestingELK-1 {
		t.Errorf("Did not read fully into buffer")
	}
	// Second piece
	p = make([]byte, 1)
	read, err = con.Read(p)
	if err != io.EOF && err != nil {
		t.Errorf(err.Error())
	}
	if read != 1 {
		t.Errorf("Did not read last byte")
	}

	// See if we can read a bit and then reset and read it all
	con.ResetReader()

	p = make([]byte, 1)
	read, err = con.Read(p)
	if read != 1 {
		t.Errorf("Did not read 1 byte")
	}
	con.ResetReader()
	// Read it all
	p = make([]byte, sizeOfTestingELK)
	read, err = con.Read(p)
	if read != sizeOfTestingELK {
		t.Errorf("Did not read entire index")
	}
	fmt.Println(string(p))
}

func TestReadLarge(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	p := make([]byte, 1024*1024)
	read, err := con.Read(p)
	fmt.Printf("Read %d bytes with error %v\n", read, err)

	// Save to file
	// ioutil.WriteFile("out.txt", p[0:read], 0755)
}
