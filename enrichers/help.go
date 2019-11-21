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
