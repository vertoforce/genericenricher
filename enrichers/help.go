package enrichers

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// stringSizeToUint Converts a string such as "3mb" to int64 byte count
func stringSizeToUint(size string) uint64 {
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
		return 0
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

// urlToIP return looked up ip from hostname
func urlToIP(url *url.URL) net.IP {
	addrs, err := net.LookupHost(url.Hostname())
	if err != nil || len(addrs) == 0 {
		return net.IP{}
	}
	return net.ParseIP(addrs[0])
}

// urlToPort return uint16 port form url
func urlToPort(url *url.URL) uint16 {
	port, err := strconv.ParseUint(url.Port(), 10, 16)
	if err != nil {
		return 80
	}

	return uint16(port)
}
