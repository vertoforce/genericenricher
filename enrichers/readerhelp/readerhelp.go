package readerhelp

import "io"

// ReaderState Struct to store position in p and source data can be put in to p over multiple function calls
type ReaderState struct {
	curEntry  []byte
	pPos      int // Position in p
	sourcePos int // Position in source
}

// PlaceStream is a helper function to help implement io.Reader by making it simple to fill a []byte given a stream of []byte arrays
func (state *ReaderState) PlaceStream(p []byte, entries chan []byte) (n int, err error) {
	read := 0

	// Keep reading into p until p is full or entries channel is empty
	for {
		// Get next entry if we need it
		if state.sourcePos == 0 {
			var ok bool
			state.curEntry, ok = <-entries
			if !ok { // EOF
				return read, io.EOF
			}
		}

		// Copy to p and see if we need to copy more
		copied, needNextEntry := state.ReadIntoP(p, state.curEntry)
		read += copied
		if !needNextEntry {
			// We are done
			return read, nil
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
