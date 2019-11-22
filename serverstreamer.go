// Package serverstreamer abstracts away different server types such as ELK and FTP servers and gives a raw stream of the data hosted on these servers.
// This raw stream can be useful to search regex rules or yara rules against.
package serverstreamer

import (
	"errors"
	"io"
	"net"
	"serverstreamer/enrichers"
)

// Server Interface to read data straight from a server
type Server interface {
	GetIP() net.IP
	GetPort() int16
	Connect() error
	IsConnected() bool
	Type() enrichers.ServerType
	io.ReadCloser
}

// GetServer Given a connection string, attempt to determine server type and return a Server
func GetServer(connectString string) (Server, error) {
	// Detect type
	serverType := enrichers.DetectServerType(connectString)
	return GetServerWithType(connectString, serverType)
}

// GetServerWithType Given a connection string and server type, return a Server
func GetServerWithType(connectString string, serverType enrichers.ServerType) (Server, error) {
	switch serverType {
	case enrichers.ELK:
		return enrichers.NewELK(connectString)
	default:
		return nil, errors.New("unknown server type")
	}
}
