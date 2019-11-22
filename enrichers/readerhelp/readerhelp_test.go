package readerhelp

import (
	"io"
	"reflect"
	"testing"
)

func TestReadIntoP(t *testing.T) {
	// Test write that ends buffer and entry, then write to a new buffer
	state := &ReaderState{}
	p := make([]byte, 1)
	p2 := make([]byte, 1)
	read, needNextSource := state.ReadIntoP(p, []byte{1})
	if read != 1 || needNextSource != false {
		t.Errorf("Incorrect return values")
	}
	read, needNextSource = state.ReadIntoP(p2, []byte{2})
	if read != 1 || needNextSource != false {
		t.Errorf("Incorrect return values")
	}
	if !reflect.DeepEqual(p, []byte{1}) {
		t.Errorf("Did not read into p correctly")
	}
	if !reflect.DeepEqual(p2, []byte{2}) {
		t.Errorf("Did not read into p correctly")
	}

	// Test single write to p with too much source
	state = &ReaderState{}
	p = make([]byte, 3)
	read, needNextSource = state.ReadIntoP(p, []byte{1, 2, 3, 4, 5})
	if read != 3 || needNextSource != false {
		t.Errorf("Incorrect return values")
	}
	if !reflect.DeepEqual(p, []byte{1, 2, 3}) {
		t.Errorf("Did not read into p correctly")
	}

	// Test single write to p with too little source
	state = &ReaderState{}
	p = make([]byte, 5)
	read, needNextSource = state.ReadIntoP(p, []byte{1, 2, 3})
	if read != 3 || needNextSource != true {
		t.Errorf("Incorrect return values")
	}
	if !reflect.DeepEqual(p, []byte{1, 2, 3, 0, 0}) {
		t.Errorf("Did not read into p correctly")
	}

	// Test two writes to p
	state = &ReaderState{}
	p = make([]byte, 10)
	read, needNextSource = state.ReadIntoP(p, []byte{1, 2, 3, 4, 5})
	if read != 5 || needNextSource != true {
		t.Errorf("Incorrect return values")
	}
	read, needNextSource = state.ReadIntoP(p, []byte{1, 2, 3, 4, 5})
	if read != 5 || needNextSource != false {
		t.Errorf("Incorrect return values")
	}
	if !reflect.DeepEqual(p, []byte{1, 2, 3, 4, 5, 1, 2, 3, 4, 5}) {
		t.Errorf("Did not read into p correctly")
	}

	// Test three writes to p
	state = &ReaderState{}
	p = make([]byte, 10)
	state.ReadIntoP(p, []byte{1, 2})
	state.ReadIntoP(p, []byte{1, 2, 3})
	state.ReadIntoP(p, []byte{1, 2, 3, 4, 5})
	if !reflect.DeepEqual(p, []byte{1, 2, 1, 2, 3, 1, 2, 3, 4, 5}) {
		t.Errorf("Did not read into p correctly")
	}

}

func TestPlaceStream(t *testing.T) {
	entries := make(chan []byte)

	// Populate entries
	go func() {
		defer close(entries)

		for i := 0; i < 3; i++ {
			entries <- []byte{1, 2, 3}
		}
		entries <- []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
		for i := 0; i < 9; i++ {
			entries <- []byte{byte(i % 3)}
		}
	}()

	// Test populating within same buffer
	state := &ReaderState{}
	p := make([]byte, 9)
	read, err := state.PlaceStream(p, entries)
	if read != 9 || err != nil {
		t.Errorf("Incorrect return values")
	}
	if !reflect.DeepEqual(p, []byte{1, 2, 3, 1, 2, 3, 1, 2, 3}) {
		t.Errorf("Did not fill stream correctly")
	}

	// Test populating within the same entry
	state = &ReaderState{}
	ps := [][]byte{make([]byte, 3), make([]byte, 3), make([]byte, 3)}
	read, err = state.PlaceStream(ps[0], entries)
	if read != 3 || err != nil {
		t.Errorf("Incorrect return values")
	}
	read, err = state.PlaceStream(ps[1], entries)
	if read != 3 || err != nil {
		t.Errorf("Incorrect return values")
	}
	read, err = state.PlaceStream(ps[2], entries)
	if read != 3 || err != nil {
		t.Errorf("Incorrect return values")
	}
	if !reflect.DeepEqual(ps, [][]byte{[]byte{0, 1, 2}, []byte{3, 4, 5}, []byte{6, 7, 8}}) {
		t.Errorf("Did not fill stream correctly")
	}

	// Test populating three times
	state = &ReaderState{}
	for i := 0; i < 3; i++ {
		p := make([]byte, 3)
		read, err = state.PlaceStream(p, entries)
		if read != 3 || err != nil {
			t.Errorf("Incorrect return values")
		}
		if !reflect.DeepEqual(p, []byte{0, 1, 2}) {
			t.Errorf("Did not fill stream correctly")
		}
	}

	// Test reading to EOF
	p = make([]byte, 3)
	state = &ReaderState{}
	read, err = state.PlaceStream(p, entries)
	if read != 0 {
		t.Errorf("Should have read 0")
	}
	if err != io.EOF {
		t.Errorf("Should have gotten EOF")
	}
}
