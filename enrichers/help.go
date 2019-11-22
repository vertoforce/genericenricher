package enrichers

import (
	"errors"
	"strconv"
	"strings"
)

// StringSizeToUint Converts a string such as "3mb" to int64 byte count
func StringSizeToUint(size string) uint64 {
	// Standardize size
	size = strings.ToLower(size)

	// Init vars
	got := float64(0)
	var err error = nil

	// Base cases
	if size == "0" {
		return 0
	}
	if len(size) < 2 {
		err = errors.New("too small")
	}

	// Parse
	switch {
	case size[len(size)-2:] == "kb":
		size = size[0 : len(size)-2]
		got, err = strconv.ParseFloat(size, 64)
		got *= 1024
	case size[len(size)-2:] == "mb":
		size = size[0 : len(size)-2]
		got, err = strconv.ParseFloat(size, 64)
		got *= 1024 * 1024
	case size[len(size)-2:] == "gb":
		size = size[0 : len(size)-2]
		got, err = strconv.ParseFloat(size, 64)
		got *= 1024 * 1024 * 1024
	case size[len(size)-2:] == "tb":
		size = size[0 : len(size)-2]
		got, err = strconv.ParseFloat(size, 64)
		got *= 1024 * 1024 * 1024 * 1024
	case size[len(size)-1:] == "b":
		size = size[0 : len(size)-1]
		got, err = strconv.ParseFloat(size, 64)
	default:
		err = errors.New("no size found")
	}

	if err != nil {
		return ^uint64(0)
	}

	return uint64(got)
}

// readIntoP Given a byte array, fill in from the source.
// returns number of bytes copied and if we need more data to fill in to p
func readIntoP(p []byte, pPos, sourcePos *int, source []byte) (n int, needNextSource bool) {
	copied := copy(p[*pPos:], source[*sourcePos:])
	*pPos += copied
	*sourcePos += copied

	if *pPos == len(p) {
		// We reached the end of this buffer, we are done
		*pPos = 0
		return copied, false
	} else if *pPos > len(p) {
		// This should never happen
		panic("copied more data into size of buf array")
	} else { //*pPos < len(p)
		// We have more to fill in here, go to next item
		*sourcePos = 0
		return copied, true
	}
}
