// Example of implementing io.Reader Read()
package readerhelp

import (
	"context"
	"fmt"
	"io"
)

// We want to implement io.Reader on this structure
type Apple struct {
	readerState *ReaderState
}

func NewApple() *Apple {
	a := &Apple{}

	// Init reader
	a.readerState = New(context.Background())

	// Get Data source (using our reader's context)
	entries := getAllData(a.readerState.ReadCtx)
	a.readerState.SetEntries(entries)

	return a
}

func Example() {
	// Create a io.Reader of our structure
	var mine io.Reader
	mine = NewApple()

	// Read from my structure
	p := make([]byte, 9)
	mine.Read(p)
	fmt.Println(p)

	// Output: [1 2 3 1 2 3 1 2 3]
}

func (a *Apple) Read(p []byte) (n int, err error) {
	return a.readerState.Read(p)
}

// getAllData Get all data in the form of a channel of bytes
func getAllData(ctx context.Context) chan []byte {
	ret := make(chan []byte)

	go func() {
		defer close(ret)

		for i := 0; i < 3; i++ {
			select {
			case ret <- []byte{1, 2, 3}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ret
}
