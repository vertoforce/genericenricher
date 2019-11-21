// Package serverstreamer abstracts away different server types such as ELK and FTP servers and gives a raw stream of the data hosted on these servers.
// This raw stream can be useful to search regex rules or yara rules against
package serverstreamer

import (
	"io"
	"net"
)

// ServerType Type of server (ELK, FTP, etc)
type ServerType int

// Server Types
const (
	Unknown ServerType = iota
	ELK
	FTP
	SSH
)

// ServerStream Interface to read data straight from a server
type ServerStream interface {
	io.ReadCloser
}

// GetStream Given an ip and server type, get a data stream
func GetStream(ip net.IP, serverType ServerType) ServerStream {
	return nil
}
