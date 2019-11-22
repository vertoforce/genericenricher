package readerhelp

import "context"

func ExampleNew() {
	// Init reader
	reader := New(context.Background())

	// Entries source
	entries := make(chan []byte)
	reader.SetEntries(entries)

	// Now we can read!
	p := make([]byte, 10)
	reader.Read(p)
}
