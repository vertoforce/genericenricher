// Package genericenricher abstracts away different server types such as ELK and FTP servers and gives a raw stream of the data hosted on these servers.
// This raw stream can be useful to search regex rules or yara rules against.
package genericenricher

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/vertoforce/genericenricher/enrichers"
)

// Server Interface to read data straight from a server
type Server interface {
	// Things to consider:
	// GetItemsMatchingRules(multiregex.RuleSet) []string

	GetIP() net.IP
	GetPort() uint16
	Connect(ctx context.Context) error
	IsConnected() bool
	Type() enrichers.ServerType
	io.ReadCloser
	ResetReader() error // Go back to start of data
}

// GetServer Given a connection string, attempt to determine server type and return a Server, if you know the server type use GetServerWithType.
func GetServer(connectString string) (Server, error) {
	// Detect type
	serverType := enrichers.DetectServerType(connectString)
	return GetServerWithType(connectString, serverType)
}

// GetServerWithType Given a connection string and server type, return a Server
// If you do not know the connectString use `enrichers.GetConnectionString`
func GetServerWithType(connectString string, serverType enrichers.ServerType) (Server, error) {
	switch serverType {
	case enrichers.ELK:
		return enrichers.NewELK(connectString)
	case enrichers.FTP:
		return enrichers.NewFTP(connectString)
	case enrichers.SQL:
		return enrichers.NewSQL(connectString)
	case enrichers.HTTP:
		return enrichers.NewHTTP(connectString)
	default:
		return nil, errors.New("unknown server type")
	}
}
