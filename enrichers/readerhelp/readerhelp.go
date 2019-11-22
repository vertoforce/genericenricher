/*
Package readerhelp is a black box like this:
	chan []byte -> with readerhelp -> implement io.Reader.Read()

Basically this helps you implement io.Reader Read() if you have a chan []byte as input.
*/
package readerhelp

import (
	"context"
	"errors"
	"io"
)

// ReaderState Struct to store position in p and source data can be put in to p over multiple function calls
type ReaderState struct {
	entries    chan []byte
	curEntry   []byte
	pPos       int // Position in p
	sourcePos  int // Position in source
	ReadCtx    context.Context
	ReadCancel context.CancelFunc
}

// New Create reader state to be ready to read.  Use nil to init non active object
func New(ctx context.Context) *ReaderState {
	state := &ReaderState{}
	state.ReadCtx, state.ReadCancel = context.WithCancel(ctx)
	state.curEntry = nil
	state.pPos = 0
	state.sourcePos = 0

	return state
}

// SetEntries sets the entries that we will read from as a stream
func (state *ReaderState) SetEntries(entries chan []byte) {
	state.sourcePos = 0
	state.entries = entries
}

// Stop Cancels readCtx and sets ReaderState to 0 state
func (state *ReaderState) Stop() {
	// Reset context
	if state.ReadCtx != nil && state.ReadCancel != nil {
		state.ReadCancel()
	}

	// Reset other
	state.entries = nil
	state.curEntry = nil
	state.pPos = 0
	state.sourcePos = 0
}

// IsActive Returns true if we are currently reading from a stream
func (state *ReaderState) IsActive() bool {
	return state.ReadCtx != nil && state.ReadCtx.Err() == nil
}

// PlaceStream is a helper function to help implement io.Reader Read function by making it simple to fill a []byte given a stream of []byte arrays
func (state *ReaderState) Read(p []byte) (n int, err error) {
	if !state.IsActive() {
		return 0, errors.New("readerstate not activated")
	}

	n = 0

	// Keep reading into p until p is full or entries channel is empty
	for {
		// Get next entry if we need it
		if state.sourcePos == 0 {
			var ok bool
			state.curEntry, ok = <-state.entries
			if !ok { // EOF
				return n, io.EOF
			}
		}

		// Copy to p and see if we need to copy more
		copied, needNextEntry := state.ReadIntoP(p, state.curEntry)
		n += copied
		if !needNextEntry {
			// We are done
			return n, nil
		}
	}
}

// ReadIntoP Given a byte array, fill in from the source.
// Returns number of bytes copied and if we need more data to fill in to p
func (state *ReaderState) ReadIntoP(p []byte, source []byte) (n int, needNextSource bool) {
	copied := copy(p[state.pPos:], source[state.sourcePos:])
	state.pPos += copied
	state.sourcePos += copied

	// Check if end of source
	if state.sourcePos == len(source) {
		state.sourcePos = 0
		needNextSource = true
	}

	// Check if end of buffer
	if state.pPos == len(p) {
		state.pPos = 0
		// If we thought we needed the next source, set to false because we actually don't
		needNextSource = false
	}

	return copied, needNextSource
}
