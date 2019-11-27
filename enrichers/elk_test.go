package enrichers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vertoforce/multiregex"
)

func TestNewElK(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// Check IP and Port
	if con.GetIP().String() != "127.0.0.1" {
		fmt.Println(con.GetIP())
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
	matchedIndices, err := con.GetIndicesMatchingRules(context.Background(), multiregex.MatchAll, 500) // Limit to 500 docs to avoid long wait times
	if err != nil {
		t.Errorf("Failed to get indices")
	}

	// Make sure we matches all indices
	if len(indices) != len(matchedIndices) {
		// Print the indices we matched
		for _, relevantIndex := range matchedIndices {
			fmt.Println(relevantIndex)
		}
		t.Errorf("Did not match all indices")
	}
}

func TestGetData(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	indices, err := con.GetIndices(context.Background())
	if len(indices) == 0 || err != nil {
		t.Errorf("No indices to test on")
	}
	testingIndex := indices[0].Index

	// Check to make sure the limit works
	ctx, cancel := context.WithCancel(context.Background())
	dataStream := con.GetData(ctx, testingIndex, 1)

	totalHits := 0
	for range dataStream {
		totalHits++
		if totalHits > 1 {
			t.Errorf("Too many hits, limit didn't work")
			break
		}
	}
	cancel()
	time.Sleep(time.Second)

	// Check cancel works
	ctx, cancel = context.WithCancel(context.Background())
	dataStream = con.GetData(ctx, testingIndex, -1)

	totalHits = 0
	for range dataStream {
		totalHits++
		cancel()
		time.Sleep(time.Millisecond * 10) // Wait for it to actually cancel

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

func TestGetTotalSize(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	_, err = con.GetTotalSize(context.Background())
	if err != nil {
		t.Errorf("Error getting size: " + err.Error())
		return
	}
}

func TestRead(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	// Check if we can read some data
	p := make([]byte, 1024)
	read, err := con.Read(p)
	if read == 0 {
		t.Errorf("Did not read anything")
	}

	// Reset and read
	con.ResetReader()
	p = make([]byte, 1024)
	read, err = con.Read(p)
	if read == 0 {
		t.Errorf("Did not read anything")
	}
}

func TestReadLarge(t *testing.T) {
	con, err := NewELK("http://127.0.0.1:9200")
	if err != nil {
		t.Errorf("failed to connect")
		return
	}

	p := make([]byte, 1024*1024)
	read, err := con.Read(p)
	if read == 0 {
		t.Errorf("Read %d bytes with error %v\n", read, err)

	}

	// Save to file
	// ioutil.WriteFile("out.txt", p[0:read], 0755)
}
